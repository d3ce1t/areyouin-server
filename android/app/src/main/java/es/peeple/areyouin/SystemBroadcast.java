package es.peeple.areyouin;

import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;

public class SystemBroadcast extends BroadcastReceiver {
    public SystemBroadcast() {
    }

    @Override
    public void onReceive(Context context, Intent intent) {

        if(intent.getAction().equalsIgnoreCase(Intent.ACTION_BOOT_COMPLETED)) {
            //here we start the service
            Intent serviceIntent = new Intent(context, AyiService.class);
            context.startService(serviceIntent);
        }
    }
}