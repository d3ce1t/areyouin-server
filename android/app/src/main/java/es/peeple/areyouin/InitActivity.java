package es.peeple.areyouin;

import android.app.Activity;
import android.content.ComponentName;
import android.content.Context;
import android.content.Intent;
import android.content.ServiceConnection;
import android.os.IBinder;
import android.os.Bundle;

import com.facebook.AccessToken;


public class InitActivity extends Activity implements ServiceConnection {

    private boolean mBound = false;
    private AyiService mService;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);

        // Starts AyiService
        Intent serviceIntent = new Intent(this, AyiService.class);
        startService(serviceIntent);

        // Bind to service
        bindService(serviceIntent, this, Context.BIND_AUTO_CREATE);
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        // Unbind from the service
        if (mBound) {
            unbindService(this);
            mBound = false;
        }
    }

    @Override
    public void onServiceConnected(ComponentName name, IBinder service) {

        AyiService.LocalBinder binder = (AyiService.LocalBinder) service;
        mService = binder.getService();
        mBound = true;

        if (!mService.isLoginNeeded()) {
            Intent intent = new Intent(this, MainActivity.class);
            startActivity(intent);
        }
        else {
            Intent intent = new Intent(this, SignInActivity.class);
            startActivity(intent);
        }

        finish();
    }

    @Override
    public void onServiceDisconnected(ComponentName name) {
        mService = null;
        mBound = false;
    }
}