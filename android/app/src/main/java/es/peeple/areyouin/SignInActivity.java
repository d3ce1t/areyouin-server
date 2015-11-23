package es.peeple.areyouin;

import android.animation.Animator;
import android.animation.AnimatorListenerAdapter;
import android.annotation.TargetApi;
import android.content.ComponentName;
import android.content.Context;
import android.content.Intent;
import android.content.ServiceConnection;
import android.os.Build;
import android.os.Bundle;
import android.os.IBinder;
import android.support.v7.app.AppCompatActivity;
import android.util.Log;
import android.view.View;
import android.view.View.OnClickListener;

import android.widget.Toast;

import com.facebook.AccessToken;
import com.facebook.CallbackManager;
import com.facebook.FacebookCallback;
import com.facebook.FacebookException;
import com.facebook.FacebookSdk;
import com.facebook.GraphRequest;
import com.facebook.GraphResponse;
import com.facebook.login.LoginManager;
import com.facebook.login.LoginResult;

import org.json.JSONException;
import org.json.JSONObject;
import java.util.Arrays;

import mehdi.sakout.fancybuttons.FancyButton;


public class SignInActivity extends AppCompatActivity {

    private static final String TAG = "SignInActivity";

    private AyiService mService;
    boolean mBound = false;

    // UI references
    private View mProgressView;
    private View mLoginFormView;
    private FancyButton mLoginButton;
    private CallbackManager mCallbackManager;

    private FacebookCallback<LoginResult> mCallback = new FacebookCallback<LoginResult>() {
        @Override
        public void onSuccess(LoginResult loginResult) {
            mLoginButton.setText(getString(R.string.login_button_sign_out));
            Toast.makeText(SignInActivity.this, "Login successful", Toast.LENGTH_SHORT).show();
            getFBAccountInfo(loginResult.getAccessToken());
        }
        @Override
        public void onCancel() {
            Toast.makeText(SignInActivity.this, "Login attempt canceled", Toast.LENGTH_SHORT).show();
            showProgress(false);
        }
        @Override
        public void onError(FacebookException e) {
            Toast.makeText(SignInActivity.this, "Login attempt failed", Toast.LENGTH_SHORT).show();
            showProgress(false);
            Log.v(TAG, e.toString());
        }
    };

    private ServiceConnection mConnection = new ServiceConnection() {
        @Override
        public void onServiceConnected(ComponentName className, IBinder service) {
            AyiService.LocalBinder binder = (AyiService.LocalBinder) service;
            mService = binder.getService();
            mBound = true;
            mLoginButton.setOnClickListener(new OnClickListener() {
                @Override
                public void onClick(View v) {
                    if (AccessToken.getCurrentAccessToken() == null) {
                        LoginManager.getInstance().logInWithReadPermissions(SignInActivity.this,
                                Arrays.asList("public_profile", "user_friends", "email"));
                        showProgress(true);
                    } else {
                        LoginManager.getInstance().logOut();
                        mLoginButton.setText(getString(R.string.login_button_sign_in));
                    }
                }
            });
        }
        @Override
        public void onServiceDisconnected(ComponentName arg0) {
            mService = null;
            mBound = false;
        }
    };

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        // Init Facebook SDK to use Facebook Login
        FacebookSdk.sdkInitialize(getApplicationContext());
        mCallbackManager = CallbackManager.Factory.create();
        LoginManager.getInstance().registerCallback(mCallbackManager, mCallback);
        LoginManager.getInstance().logOut();

        setContentView(R.layout.activity_signin);

        mProgressView = findViewById(R.id.login_progress);
        mLoginFormView = findViewById(R.id.login_form);
        mLoginButton = (FancyButton) findViewById(R.id.login_button);

        // Setup custom Facebook button
        if (AccessToken.getCurrentAccessToken() != null) {
            mLoginButton.setText(getString(R.string.login_button_sign_out));
        }
    }

    @Override
    protected void onStart() {
        super.onStart();
        // Bind to AyiService
        Intent intent = new Intent(this, AyiService.class);
        bindService(intent, mConnection, Context.BIND_AUTO_CREATE);
    }

    @Override
    protected void onStop() {
        super.onStop();
        // Unbind from the service
        if (mBound) {
            unbindService(mConnection);
            mBound = false;
        }
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        mCallbackManager.onActivityResult(requestCode, resultCode, data);

        if (requestCode == 1) {
            if (resultCode == RESULT_OK) {
                doUserAuthentication();
            } else {
                LoginManager.getInstance().logOut();
                mLoginButton.setText(getString(R.string.login_button_sign_in));
                Toast.makeText(SignInActivity.this, "Operation cancelled", Toast.LENGTH_SHORT).show();
            }
        }
    }

    private void getFBAccountInfo(final AccessToken token) {

        GraphRequest request = GraphRequest.newMeRequest(
                token,
                new GraphRequest.GraphJSONObjectCallback() {
                    @Override
                    public void onCompleted(JSONObject object, GraphResponse response) {
                        if (object != null) {
                            attemptGetAccessToken(object, token);
                        } else {
                            showProgress(false);
                        }
                    }
                });

        Bundle parameters = new Bundle();
        parameters.putString("fields", "id,name,email");
        request.setParameters(parameters);
        request.executeAsync();
    }

    private void attemptGetAccessToken(final JSONObject fbInfo, final AccessToken token) {

        if (mService == null) {
            return;
        }

        if (!mService.isConnected()) {
            Toast.makeText(SignInActivity.this, "Service isn't connected", Toast.LENGTH_SHORT).show();
            return;
        }

        mService.newAuthTokenByFacebook(token.getUserId(), token.getToken(), new AyiService.ResultListener() {
            @Override
            public void result(int error) {

                if (error == AyiService.E_NO_ERROR) {
                    Toast.makeText(SignInActivity.this, "Access granted", Toast.LENGTH_SHORT).show();
                    doUserAuthentication();
                }
                else if (error == AyiService.E_INVALID_USER) {
                    try {
                        // Create a new account
                        Intent intent = new Intent(SignInActivity.this, SignUpActivity.class);
                        intent.putExtra("fbid", token.getUserId());
                        intent.putExtra("fbtoken", token.getToken());

                        if (fbInfo.has("name")) {
                            intent.putExtra("name", fbInfo.getString("name"));
                        }

                        if (fbInfo.has("email")) {
                            intent.putExtra("email", fbInfo.getString("email"));
                        }

                        startActivityForResult(intent, 1);
                        overridePendingTransition(R.anim.pull_in_right, R.anim.push_out_left);

                    } catch (JSONException ex) {
                        Toast.makeText(SignInActivity.this, "Unknown error (1)", Toast.LENGTH_SHORT).show();
                    }
                    finally {
                        showProgress(false);
                    }
                } else {
                    Toast.makeText(SignInActivity.this, "Unknown error (2)", Toast.LENGTH_SHORT).show();
                    showProgress(false);
                }
            }
        });
    }

    private void doUserAuthentication() {
        mService.doUserAuthentication(new AyiService.ResultListener() {
            @Override
            public void result(int error) {
                if (error == AyiService.E_NO_ERROR) {
                    Intent intent = new Intent(SignInActivity.this, MainActivity.class);
                    startActivity(intent);
                    finish();
                } else {
                    Toast.makeText(SignInActivity.this, "An error occurred while trying to login", Toast.LENGTH_SHORT).show();
                }
                showProgress(false);
            }
        });
    }

    /*private boolean isEmailValid(String email) {
        //TODO: Replace this with your own logic
        return email.contains("@");
    }

    private boolean isPasswordValid(String password) {
        //TODO: Replace this with your own logic
        return password.length() > 4;
    }*/

    /**
     * Shows the progress UI and hides the login form.
     */
    @TargetApi(Build.VERSION_CODES.HONEYCOMB_MR2)
    private void showProgress(final boolean show) {
        // On Honeycomb MR2 we have the ViewPropertyAnimator APIs, which allow
        // for very easy animations. If available, use these APIs to fade-in
        // the progress spinner.
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.HONEYCOMB_MR2) {
            int shortAnimTime = getResources().getInteger(android.R.integer.config_shortAnimTime);

            mLoginFormView.setVisibility(show ? View.GONE : View.VISIBLE);
            mLoginFormView.animate().setDuration(shortAnimTime).alpha(
                    show ? 0 : 1).setListener(new AnimatorListenerAdapter() {
                @Override
                public void onAnimationEnd(Animator animation) {
                    mLoginFormView.setVisibility(show ? View.GONE : View.VISIBLE);
                }
            });

            mProgressView.setVisibility(show ? View.VISIBLE : View.GONE);
            mProgressView.animate().setDuration(shortAnimTime).alpha(
                    show ? 1 : 0).setListener(new AnimatorListenerAdapter() {
                @Override
                public void onAnimationEnd(Animator animation) {
                    mProgressView.setVisibility(show ? View.VISIBLE : View.GONE);
                }
            });
        } else {
            // The ViewPropertyAnimator APIs are not available, so simply show
            // and hide the relevant UI components.
            mProgressView.setVisibility(show ? View.VISIBLE : View.GONE);
            mLoginFormView.setVisibility(show ? View.GONE : View.VISIBLE);
        }
    }
}