package es.peeple.areyouin;

import android.content.ComponentName;
import android.content.ServiceConnection;
import android.os.IBinder;

public class AyiServiceConnection implements ServiceConnection {

    private boolean mBound = false;
    private AyiService mService;

    @Override
    public void onServiceConnected(ComponentName className, IBinder service) {
        AyiService.LocalBinder binder = (AyiService.LocalBinder) service;
        mService = binder.getService();
        mBound = true;
    }

    @Override
    public void onServiceDisconnected(ComponentName arg0) {
        mService = null;
        mBound = false;
    }

    public boolean isBound() {
        return mBound;
    }

    public AyiService getService() {
        return mService;
    }
}
