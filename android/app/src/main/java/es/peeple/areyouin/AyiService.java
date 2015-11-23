package es.peeple.areyouin;

import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.Context;
import android.content.Intent;
import android.content.SharedPreferences;
import android.os.Binder;
import android.os.Handler;
import android.os.IBinder;
import android.support.v4.app.NotificationCompat;
import android.util.Log;

import com.facebook.AccessToken;
import com.facebook.FacebookSdk;

import java.io.EOFException;
import java.io.IOException;
import java.io.InputStream;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.util.UUID;

import es.peeple.areyouin.protocol.AyiHeader;
import es.peeple.areyouin.protocol.AyiPacket;
import es.peeple.areyouin.protocol.Protocol;

// TODO: Consider when e-mail permission is revoked or friends list is revoked
// TODO: Consider what To Do when connection is lost

/**
 * AyiService to manage user communication with AyiServer
 */
public class AyiService extends Service {

    // Logger
    private final static String TAG = AyiService.class.getSimpleName();

    // Errors definitions
    public final static int E_NO_ERROR = 0;
    public final static int E_INVALID_USER = 1;
    public final static int E_USER_EXISTS = 2;
    public final static int E_FB_MISSING_DATA = 3;
    public final static int E_FB_INVALID_TOKEN = 4;
    public final static int E_MALFORMED_MESSAGE = 5;

    // My Application Errors
    public final static int E_CLOSED_CONNECTION = 6;
    public final static int E_NO_LOGIN_DATA = 7;
    public final static int E_IO_UNEXPECTED_ERROR = 8;

    // Ok definitions
    private final static int OK_AUTH = 0;

    // Preferences
    private final static String PREFS_NAME = "Preferences";
    private final static String F_USER_ID = "UserId";
    private final static String F_AUTH_TOKEN = "AuthToken";

    // Notifications
    private final static int NOTIFY_LOGIN = 1;
    private final static int NOTIFY_SERVER = 2;

    // Internet
    private final static String HOST = "192.168.1.3";
    private final static int SERVER_PORT = 1822;

    // Class objects
    private ReceiverThread mReceiver;
    private Handler mResponseHandler = new Handler(); // It's created on UI main thread
    private Socket mSocket;

    // Status variables
    private boolean mAuthenticated = false;
    private ResultListener mCreateAccountListener = null;
    private ResultListener mNewAuthTokenListener = null;
    private ResultListener mAuthListener = null;
    private boolean mConnecting = false;

    // System service
    private NotificationManager mNotificationManager;

    // Facebook Login
    private AccessToken mAccessToken;

    // Binder given to clients
    private final IBinder mBinder = new LocalBinder();


    public class LocalBinder extends Binder {
        public AyiService getService() {
            return AyiService.this;
        }
    }

    @Override
    public IBinder onBind(Intent intent) {
        return mBinder;
    }

    @Override
    public void onCreate() {
        super.onCreate();

        // Init Facebook SDK
        FacebookSdk.sdkInitialize(getApplicationContext());

        mNotificationManager =  (NotificationManager) getSystemService(Context.NOTIFICATION_SERVICE);

        // Restore preferences
        SharedPreferences settings = getSharedPreferences(PREFS_NAME, 0);
        String userId = settings.getString(F_USER_ID, null);
        String authToken = settings.getString(F_AUTH_TOKEN, null);

        if (userId == null || authToken == null) {
            asyncConnect(new AsyncConnectListener() {
                @Override
                public void onFinished(IOException error) {
                    if (error == null) {
                        notifyAuthNeeded();
                    }
                }
            });
        } else {
            connectService(userId, authToken);
        }
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        super.onStartCommand(intent, flags, startId);
        return START_STICKY; // FIXME: Reboot service when killed has no effect when stopped from Settins->Applications menu
    }

    @Override
    public void onDestroy() {
        close();
    }

    public interface AsyncConnectListener {
        void onFinished(IOException error);
    }

    public interface ResultListener {
        void result(int error_code);
    }

    /** Method for clients */
    public boolean isAuthenticated() {
        return mAuthenticated;
    }

    public boolean isConnected() {
        return mSocket != null && mSocket.isConnected();
    }

    public boolean isLoginNeeded() {
        // Assume F_USER_ID and F_AUTH_TOKEN are valid tokens
        SharedPreferences settings = getSharedPreferences(PREFS_NAME, 0);
        String userId = settings.getString(F_USER_ID, null);
        String authToken = settings.getString(F_AUTH_TOKEN, null);
        boolean firstLogin = false;

        if (userId == null || authToken == null) {
            firstLogin = true;
        }

        return firstLogin;
    }

    public void createEvent() throws IOException {
        checkService();
    }

    public void cancelEvent() throws IOException {
        checkService();
    }

    public void inviteUsers() throws IOException {
        checkService();
    }

    public void cancelUsersInvitation() throws IOException {
        checkService();
    }

    public void confirmAttendance() throws IOException {
        checkService();
    }

    public void modifyEventDate() throws IOException {
        checkService();
    }

    public void modifyEventMessage() throws IOException {
        checkService();
    }

    public void modifyEvent() throws IOException {
        checkService();
    }

    public void voteChange() throws IOException {
        checkService();
    }

    public void userPosition() throws IOException {
        checkService();
    }

    public void userPositionRange() throws IOException {
        checkService();
    }

    public void createUserAccount(String name, String phone, String email, String password, String fbid, String fbtoken, ResultListener listener) {

        if (!isConnected()) {
            listener.result(E_CLOSED_CONNECTION);
            return;
        }

        if (mCreateAccountListener != null)
            return;

        Protocol.CreateUserAccount createAccountMsg =
                Protocol.CreateUserAccount.newBuilder()
                        .setName(name)
                        .setPhone(phone)
                        .setEmail(email)
                        .setPassword(password)
                        .setFbid(fbid)
                        .setFbtoken(fbtoken)
                        .build();

        try {
            AyiPacket packet = AyiPacket.newPacket(AyiHeader.M_USER_CREATE_ACCOUNT, createAccountMsg);
            packet.writeTo(mSocket.getOutputStream());
            mCreateAccountListener = listener;
        }
        catch (IOException ex) {
            listener.result(E_IO_UNEXPECTED_ERROR);
        }
    }

    public void newAuthTokenByEmail(String email, String password, ResultListener listener) {

        if (!isConnected()) {
            listener.result(E_CLOSED_CONNECTION);
            return;
        }

        if (mNewAuthTokenListener != null)
            return;

        clearAuthTokens();
        mAuthenticated = false;

        Protocol.NewAuthToken newTokenMsg =
                Protocol.NewAuthToken.newBuilder()
                    .setPass1(email)
                    .setPass2(password)
                    .setType(Protocol.AuthType.A_NATIVE)
                    .build();

        try {
            AyiPacket packet = AyiPacket.newPacket(AyiHeader.M_USER_NEW_AUTH_TOKEN, newTokenMsg);
            packet.writeTo(mSocket.getOutputStream());
            mNewAuthTokenListener = listener;
        }
        catch (IOException ex) {
            listener.result(E_IO_UNEXPECTED_ERROR);
        }
    }

    public void newAuthTokenByFacebook(String fbId, String fbToken, ResultListener listener) {

        if (!isConnected()) {
            listener.result(E_CLOSED_CONNECTION);
            return;
        }

        if (mNewAuthTokenListener != null)
            return;

        clearAuthTokens();
        mAuthenticated = false;

        Protocol.NewAuthToken newTokenMsg =
                Protocol.NewAuthToken.newBuilder()
                        .setPass1(fbId)
                        .setPass2(fbToken)
                        .setType(Protocol.AuthType.A_FACEBOOK)
                        .build();

        try {
            AyiPacket packet = AyiPacket.newPacket(AyiHeader.M_USER_NEW_AUTH_TOKEN, newTokenMsg);
            packet.writeTo(mSocket.getOutputStream());
            mNewAuthTokenListener = listener;
        }
        catch (IOException ex) {
            listener.result(E_IO_UNEXPECTED_ERROR);
        }
    }

    public void doUserAuthentication(ResultListener listener) {

        if (!isConnected()) {
            listener.result(E_CLOSED_CONNECTION);
            return;
        }

        if (mAuthListener != null)
            return;

        // Restore preferences
        SharedPreferences settings = getSharedPreferences(PREFS_NAME, 0);
        String userId = settings.getString(F_USER_ID, null);
        String authToken = settings.getString(F_AUTH_TOKEN, null);

        if (userId == null || authToken == null) {
            listener.result(E_NO_LOGIN_DATA);
            return;
        }

        try {
            userAuthentication(userId, authToken);
            mAuthListener = listener;
        } catch (IOException e) {
            listener.result(E_IO_UNEXPECTED_ERROR);
        }
    }

    public void ping() throws IOException {

        checkService();

        Protocol.Ping msg =
                Protocol.Ping.newBuilder()
                        .setCurrentTime(System.currentTimeMillis() / 1000)
                        .build();

        AyiPacket packet = AyiPacket.newPacket(AyiHeader.M_PING, msg);
        packet.writeTo(mSocket.getOutputStream());
    }

    public void readEvent() throws IOException {
        checkService();
    }

    public void listAuthoredEvents() throws IOException {
        checkService();
    }

    public void listPrivateEvents() throws IOException {
        checkService();
    }

    public void listPublicEvents() throws IOException {
        checkService();
    }

    public void historyAuthoredEvents() throws IOException {
        checkService();
    }

    public void historyPrivateEvents() throws IOException {
        checkService();
    }

    public void historyPublicEvents() throws IOException {
        checkService();
    }

    /** Private methods **/
    private void checkService() throws IOException {
        checkConnected();
        if (!isAuthenticated())
            throw new IOException("User isn't authenticated");
    }

    private void checkConnected() throws IOException {
        if (!isConnected())
            throw new IOException("AyiService isn't connected");
    }

    private void connectService(final String userId, final String authToken) {

        asyncConnect(new AsyncConnectListener() {
            @Override
            public void onFinished(IOException error) {
                if (error == null) {
                    // Identify with the service
                    try {
                        userAuthentication(userId, authToken);
                    } catch (IOException ex) {
                        Log.e(TAG, ex.toString());
                    }
                }
            }
        });
    }

    // FIXME: Sometimes this notification is shown along ServerDownNotification. Keep only the last one.
    private void notifyAuthNeeded() {
        NotificationCompat.Builder mBuilder =
                new NotificationCompat.Builder(this)
                        .setSmallIcon(R.mipmap.ic_launcher)
                        .setContentTitle(getString(R.string.app_name))
                        .setContentText(getString(R.string.notify_login_required))
                        .setAutoCancel(false)
                        .setOngoing(true);

        // Creates an explicit intent for an Activity in your app
        Intent intent = new Intent(this, SignInActivity.class);
        PendingIntent pendingIntent = PendingIntent.getActivity(this, 0, intent, PendingIntent.FLAG_UPDATE_CURRENT);
        mBuilder.setContentIntent(pendingIntent);
        // mId allows you to update the notification later on.
        mNotificationManager.notify(NOTIFY_LOGIN, mBuilder.build());
    }

    private void notifyServerDown() {
        NotificationCompat.Builder mBuilder =
                new NotificationCompat.Builder(this)
                        .setSmallIcon(R.mipmap.ic_launcher)
                        .setContentTitle(getString(R.string.app_name))
                        .setContentText(getString(R.string.notify_server_offline))
                        .setAutoCancel(false)
                        .setOngoing(true);

        // Creates an explicit intent for an Activity in your app
        //Intent intent = new Intent(this, SignUpActivity.class);
        //PendingIntent pendingIntent = PendingIntent.getActivity(this, 0, intent, PendingIntent.FLAG_UPDATE_CURRENT);
        //mBuilder.setContentIntent(pendingIntent);
        // mId allows you to update the notification later on.
        mNotificationManager.notify(NOTIFY_SERVER, mBuilder.build());
    }

    private void asyncConnect(final AsyncConnectListener callback) {

        if (isConnected()) {
            if (callback != null)
                callback.onFinished(new IOException("AyiService already connected"));
            Log.e(TAG, "AyiService already connected");
            return;
        }

        boolean skip_call = false;

        synchronized (this) {
            if (mConnecting) { // Skip call if there is a pending connection
                skip_call = true;
            }
            mConnecting = true;
        }

        if (skip_call) {
            if (callback != null)
                callback.onFinished(new IOException("A pending asyncConnect() call have not finalised yet"));
            Log.e(TAG, "A pending asyncConnect() call have not finalised yet");
            return;
        }

        new Thread(new Runnable() {
            @Override
            public void run() {
                IOException error = null;

                try {
                    mSocket = new Socket();
                    mSocket.connect(new InetSocketAddress(HOST, SERVER_PORT));
                    if (mReceiver == null) {
                        mReceiver = new ReceiverThread();
                    }
                    mReceiver.start();
                    mNotificationManager.cancel(NOTIFY_SERVER);
                }
                catch (IOException ex) {
                    notifyServerDown();
                    error = ex;
                    Log.e(TAG, ex.toString());
                }
                finally {
                    synchronized (AyiService.this) {
                        mConnecting = false;
                    }
                }

                final IOException result = error;
                mResponseHandler.post(new Runnable() {
                    @Override
                    public void run() {
                        if (callback != null)
                            callback.onFinished(result);
                    }
                });

            }
        }).start();
    }

    private void close() {
        try {
            if (mSocket != null) {
                mSocket.close();
                mSocket = null;
            }
        }
        catch (IOException ex) {
            Log.e(TAG, ex.toString());
        }
    }

    private void onClosedConnection() {
        try {
            mReceiver = null;
            mSocket.close();
            mSocket = null;
            finishPendingCallbacks();

        }
        catch (IOException ex) {
            Log.e(TAG, ex.toString());
        }
    }

    private void finishPendingCallbacks() {

        if (mCreateAccountListener != null) {
            mCreateAccountListener.result(E_CLOSED_CONNECTION);
            mCreateAccountListener = null;
        }

        if (mNewAuthTokenListener != null) {
            mNewAuthTokenListener.result(E_CLOSED_CONNECTION);
            mNewAuthTokenListener = null;
        }
    }

    private void userAuthentication(String user_id, String auth_token) throws IOException {
        Protocol.UserAuthentication msg =
                Protocol.UserAuthentication.newBuilder()
                        .setUserId(user_id)
                        .setAuthToken(auth_token)
                        .build();

        AyiPacket packet = AyiPacket.newPacket(AyiHeader.M_USER_AUTH, msg);
        packet.writeTo(mSocket.getOutputStream());
    }

    private void manageException(Exception ex) {
        Log.e(TAG, ex.toString());
    }

    private void onAccessGranted(UUID user_id, UUID auth_token) {
        SharedPreferences settings = getSharedPreferences(PREFS_NAME, 0);
        SharedPreferences.Editor editor = settings.edit();
        editor.putString(F_USER_ID, user_id.toString());
        editor.putString(F_AUTH_TOKEN, auth_token.toString());
        editor.commit();

        if (mCreateAccountListener != null) {
            mCreateAccountListener.result(E_NO_ERROR);
            mCreateAccountListener = null;
        }

        if (mNewAuthTokenListener != null) {
            mNewAuthTokenListener.result(E_NO_ERROR);
            mNewAuthTokenListener = null;
        }
    }

    private void processPacket(AyiPacket packet)  {

        byte[] data = packet.getData();

        try {
            switch (packet.getHeader().type) {
                // Responses
                case AyiHeader.M_PONG:
                    //mCallback.onPong( Protocol.Pong.parseFrom(data) );
                    break;

                case AyiHeader.M_EVENT_INFO:
                    //mCallback.onEventInfo(Protocol.EventInfo.parseFrom(data));
                    break;

                case AyiHeader.M_EVENTS_LIST:
                    //message = Protocol.EventsList.parseFrom(data);
                    break;

                // Notifications
                case AyiHeader.M_EVENT_CREATED:
                    //message = Protocol.EventCreated.parseFrom(data);
                    break;

                case AyiHeader.M_EVENT_CANCELLED:
                    //message = Protocol.EventCancelled.parseFrom(data);
                    break;

                case AyiHeader.M_EVENT_EXPIRED:
                    //message = Protocol.EventExpired.parseFrom(data);
                    break;

                case AyiHeader.M_EVENT_DATE_MODIFIED:
                case AyiHeader.M_EVENT_MESSAGE_MODIFIED:
                case AyiHeader.M_EVENT_MODIFIED:
                    //message = Protocol.EventModified.parseFrom(data);
                    break;

                case AyiHeader.M_INVITATION_RECEIVED:
                    //message = Protocol.InvitationReceived.parseFrom(data);
                    break;

                case AyiHeader.M_INVITATION_CANCELLED:
                    //message = Protocol.InvitationCancelled.parseFrom(data);
                    break;

                case AyiHeader.M_ATTENDANCE_STATUS:
                    //message = Protocol.AttendanceStatus.parseFrom(data);
                    break;

                case AyiHeader.M_EVENT_CHANGE_DATE_PROPOSED:
                case AyiHeader.M_EVENT_CHANGE_MESSAGE_PROPOSED:
                case AyiHeader.M_EVENT_CHANGE_PROPOSED:
                    //message = Protocol.EventChangeProposed.parseFrom(data);
                    break;

                case AyiHeader.M_VOTING_STATUS:
                    //message = Protocol.VotingStatus.parseFrom(data);
                    break;

                case AyiHeader.M_VOTING_FINISHED:
                    //message = Protocol.VotingStatus.parseFrom(data);
                    break;

                case AyiHeader.M_CHANGE_ACCEPTED:
                    //message = Protocol.ChangeAccepted.parseFrom(data);
                    break;

                case AyiHeader.M_CHANGE_DISCARDED:
                    //message = Protocol.ChangeDiscarded.parseFrom(data);
                    break;

                case AyiHeader.M_ACCESS_GRANTED:
                    Protocol.AccessGranted accessGranted = Protocol.AccessGranted.parseFrom(data);
                    UUID userId = UUID.fromString(accessGranted.getUserId());
                    UUID authToken = UUID.fromString(accessGranted.getAuthToken());
                    onAccessGranted(userId, authToken);
                    break;

                case AyiHeader.M_OK:
                    Log.v(TAG, "OK");
                    Protocol.Ok okMsg = Protocol.Ok.parseFrom(data);
                    if (okMsg.getType() == OK_AUTH) {
                        mAuthenticated = true;
                        mNotificationManager.cancel(NOTIFY_LOGIN);
                        if (mAuthListener != null) {
                            mAuthListener.result(E_NO_ERROR);
                            mAuthListener = null;
                        }
                    }
                    break;

                case AyiHeader.M_ERROR:
                    Log.v(TAG, "ERROR");
                    Protocol.Error error = Protocol.Error.parseFrom(data);
                    manageMsgErrors(error);
                    break;

                // Requests
                /*case AyiHeader.M_PING:
                    //message = Protocol.Ping.parseFrom(data);
                    break;*/

                // Modifiers
                /*case AyiHeader.M_USER_AUTH_BY_USERPASSWORD:
                case AyiHeader.M_USER_AUTH_BY_TOKEN:
                    //message = Protocol.UserAuthentication.parseFrom(data);
                    break;*/
            }
        }
        catch (IOException ex) {
            manageException(ex);
        }
    }

    private void clearAuthTokens() {
        SharedPreferences settings = getSharedPreferences(PREFS_NAME, 0);
        SharedPreferences.Editor editor = settings.edit();
        editor.remove(F_USER_ID);
        editor.remove(F_AUTH_TOKEN);
        editor.commit();
    }

    private void manageMsgErrors(Protocol.Error error) {
        switch (error.getType()) {

            case AyiHeader.M_USER_AUTH:
                clearAuthTokens();
                mAuthenticated = false;
                notifyAuthNeeded();
                if (mAuthListener != null) {
                    mAuthListener.result(error.getError());
                    mAuthListener = null;
                }
                break;

            case AyiHeader.M_USER_NEW_AUTH_TOKEN:
                if (mNewAuthTokenListener != null) {
                    mNewAuthTokenListener.result(error.getError());
                    mNewAuthTokenListener = null;
                }
                break;

            case AyiHeader.M_USER_CREATE_ACCOUNT:
                if (mCreateAccountListener != null) {
                    mCreateAccountListener.result(error.getError());
                    mCreateAccountListener = null;
                }
                break;
        }
    }

    private class ReceiverThread extends Thread {
        @Override
        public void run() {

            InputStream input;

            try {
                input = mSocket.getInputStream();
            }
            catch (final IOException ex) {
                mResponseHandler.post(new Runnable() {
                    @Override
                    public void run() {
                        manageException(ex);
                    }
                });
                return;
            }

            // Loop through packets received
            for (;;) {
                try {
                    final AyiPacket packet = AyiPacket.readFrom(input); // FIXME: I'm creating packet objects every time
                    Log.v(TAG, "Packet " + packet.getHeader().type + " Read of Size " + packet.getHeader().size);
                    mResponseHandler.post(new Runnable() {
                        @Override
                        public void run() { processPacket(packet); }
                    });
                }
                // Connection closed
                catch (EOFException ex) {
                    Log.e(TAG, ex.toString());
                    mResponseHandler.post(new Runnable() {
                        @Override
                        public void run() { onClosedConnection(); }
                    });
                    return;
                }
                catch (final IOException ex) {
                    Log.e(TAG, ex.toString());
                    mResponseHandler.post(new Runnable() {
                        @Override
                        public void run() {manageException(ex);
                        }
                    });
                }
            }
        }
    }
}
