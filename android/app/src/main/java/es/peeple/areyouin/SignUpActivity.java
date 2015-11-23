package es.peeple.areyouin;

import android.animation.Animator;
import android.animation.AnimatorListenerAdapter;
import android.annotation.TargetApi;
import android.content.ComponentName;
import android.content.Context;
import android.content.Intent;
import android.content.ServiceConnection;
import android.os.IBinder;
import android.support.v7.app.AppCompatActivity;

import android.os.Build;
import android.os.Bundle;

import android.text.TextUtils;
import android.view.View;
import android.view.View.OnClickListener;
import android.widget.Button;
import android.widget.EditText;
import android.widget.Toast;


public class SignUpActivity extends AppCompatActivity {

    private static final String TAG = "SignUpActivity";

    private AyiService mService;
    boolean mBound = false;

    private String mFbId = "";
    private String mFbToken = "";

    // UI references.
    private EditText mNameView;
    //private EditText mPhoneView;
    private EditText mEmailView;
    //private EditText mPasswordView;
    //private EditText mConfirmPasswordView;
    private View mProgressView;
    private View mLoginFormView;

    private ServiceConnection mConnection = new ServiceConnection() {
        @Override
        public void onServiceConnected(ComponentName className, IBinder service) {
            AyiService.LocalBinder binder = (AyiService.LocalBinder) service;
            mService = binder.getService();
            mBound = true;
            if (mFbId != null && mFbToken != null && !mFbId.isEmpty() && !mFbToken.isEmpty()) {
                attemptLogin();
            }
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
        setContentView(R.layout.activity_signup);

        Intent myIntent = getIntent(); // gets the previously created intent

        mFbId = myIntent.getStringExtra("fbid");
        mFbToken = myIntent.getStringExtra("fbtoken");

        mNameView = (EditText) findViewById(R.id.name);
        mNameView.setText(myIntent.getStringExtra("name"));

        //mPhoneView = (EditText) findViewById(R.id.phone);

        mEmailView = (EditText) findViewById(R.id.email);
        mEmailView.setText(myIntent.getStringExtra("email"));


        /*mPasswordView = (EditText) findViewById(R.id.password);
        mPasswordView.setOnEditorActionListener(new TextView.OnEditorActionListener() {
            @Override
            public boolean onEditorAction(TextView textView, int id, KeyEvent keyEvent) {
                if (id == EditorInfo.IME_ACTION_DONE || id == EditorInfo.IME_NULL) {
                    attemptLogin();
                    return true;
                }
                return false;
            }
        });*/

        //mConfirmPasswordView = (EditText) findViewById(R.id.confirm_password);
        Button mEmailSignInButton = (Button) findViewById(R.id.email_sign_in_button);
        mEmailSignInButton.setOnClickListener(new OnClickListener() {
            @Override
            public void onClick(View view) {
                attemptLogin();
            }
        });

        mLoginFormView = findViewById(R.id.login_form);
        mProgressView = findViewById(R.id.login_progress);
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
    public void onBackPressed() {
        super.onBackPressed();
        overridePendingTransition(R.anim.pull_in_left, R.anim.push_out_right);
    }

    /**
     * Attempts to sign in or register the account specified by the login form.
     * If there are form errors (invalid email, missing fields, etc.), the
     * errors are presented and no actual login attempt is made.
     */
    private void attemptLogin() {

        if (mService == null) {
            return;
        }

        // Reset errors.
        mNameView.setError(null);
        //mPhoneView.setError(null);
        mEmailView.setError(null);
        //mPasswordView.setError(null);
        //mConfirmPasswordView.setError(null);

        // Store values at the time of the login attempt.
        String name = mNameView.getText().toString();
        //String phone = mPhoneView.getText().toString();
        String email = mEmailView.getText().toString();
        //String password = mPasswordView.getText().toString();
        //String confirmPassword = mConfirmPasswordView.getText().toString();

        boolean cancel = false;
        View focusView = null;

        // Check for a valid confirm password
        /*if (TextUtils.isEmpty(confirmPassword)) {
            mConfirmPasswordView.setError(getString(R.string.error_field_required));
            focusView = mConfirmPasswordView;
            cancel = true;
        }
        else if (!password.equals(confirmPassword)) {
            mConfirmPasswordView.setError(getString(R.string.error_mismatch_password));
            focusView = mConfirmPasswordView;
            cancel = true;
        }*/

        // Check for a valid password
        /*if (TextUtils.isEmpty(password)) {
            mPasswordView.setError(getString(R.string.error_field_required));
            focusView = mPasswordView;
            cancel = true;
        }
        else if (!isPasswordValid(password)) {
            mPasswordView.setError(getString(R.string.error_invalid_password));
            focusView = mPasswordView;
            cancel = true;
        }*/

        // Check email address.
        if (TextUtils.isEmpty(email)) {
            mEmailView.setError(getString(R.string.error_field_required));
            focusView = mEmailView;
            cancel = true;
        } else if (!isEmailValid(email)) {
            mEmailView.setError(getString(R.string.error_invalid_email));
            focusView = mEmailView;
            cancel = true;
        }

        // Check phone number
        /*if (TextUtils.isEmpty(phone) ) {
            mPhoneView.setError(getString(R.string.error_field_required));
            focusView = mPhoneView;
            cancel = true;
        }
        else if (!isPhoneValid(phone)) {
            mPhoneView.setError(getString(R.string.error_invalid_phone));
            focusView = mPhoneView;
            cancel = true;
        }*/

        // Check name
        if (TextUtils.isEmpty(name)) {
            mNameView.setError(getString(R.string.error_field_required));
            focusView = mNameView;
            cancel = true;
        }
        else if (!isNameValid(name)) {
            mNameView.setError(getString(R.string.error_invalid_name));
            focusView = mNameView;
            cancel = true;
        }

        if (cancel) {
            // There was an error; don't attempt login and focus the first
            // form field with an error.
            focusView.requestFocus();
            return;
        }

        if (mService.isConnected()) {

            showProgress(true);

            mService.createUserAccount(name, "", email, "", mFbId, mFbToken, new AyiService.ResultListener() {
                @Override
                public void result(int error) {
                    showProgress(false);
                    if (error == AyiService.E_NO_ERROR) {
                        Toast.makeText(SignUpActivity.this, "Account created successfully", Toast.LENGTH_SHORT).show();
                        Intent returnIntent = new Intent();
                        setResult(RESULT_OK,returnIntent);
                        finish();
                    } else if (error == AyiService.E_USER_EXISTS){
                        mEmailView.setError(getString(R.string.error_user_exists));
                        mEmailView.requestFocus();
                    } else if (error == AyiService.E_FB_INVALID_TOKEN) {
                        Toast.makeText(SignUpActivity.this, "Facebook account is not accessible. Try again.", Toast.LENGTH_SHORT).show();
                    }
                    else {
                        Toast.makeText(SignUpActivity.this, "Account couldn't be created. Try again", Toast.LENGTH_SHORT).show();
                    }
                }
            });
        }
        else {
            Toast.makeText(SignUpActivity.this, "Service isn't connected", Toast.LENGTH_SHORT).show();
        }
    }

    private boolean isNameValid(String name) {
        return name.matches("[a-zA-Z0-9\\s]+") && name.length() > 4;
    }

    /*private boolean isPhoneValid(String phone) {
        return phone.matches("[0-9]+") && phone.length() <= 15;
    }*/

    private boolean isEmailValid(String email) {
        //TODO: Replace this with your own logic
        return email.contains("@");
    }

    /*private boolean isPasswordValid(String password) {
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