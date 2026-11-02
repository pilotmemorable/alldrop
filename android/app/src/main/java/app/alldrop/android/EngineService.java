package app.alldrop.android;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.PendingIntent;
import android.app.Service;
import android.content.BroadcastReceiver;
import android.content.Context;
import android.content.Intent;
import android.content.IntentFilter;
import android.content.pm.ServiceInfo;
import android.net.ConnectivityManager;
import android.net.wifi.WifiManager;
import android.os.Build;
import android.os.Handler;
import android.os.IBinder;
import android.os.Looper;
import android.os.PowerManager;

/** Keeps the engine alive in the background so the phone can receive files. */
public class EngineService extends Service {
    static final String ACTION_STOP = "app.alldrop.android.STOP";
    private static final String CH_SERVICE = "service";
    private static final String CH_OFFERS = "offers";
    private static final int ID_RUNNING = 1;
    private static final int ID_OFFER = 2;

    private static EngineService instance;
    private static final Handler main = new Handler(Looper.getMainLooper());

    private WifiManager.MulticastLock multicastLock;
    private WifiManager.WifiLock wifiLock;
    private PowerManager.WakeLock cpuLock;
    private BroadcastReceiver netReceiver;
    private boolean stopping;
    private long lastRestart;

    private final Runnable netCheck = new Runnable() {
        @Override
        public void run() {
            Engine.networkChanged(EngineService.this);
        }
    };

    static void start(Context ctx) {
        Intent i = new Intent(ctx, EngineService.class);
        if (Build.VERSION.SDK_INT >= 26) ctx.startForegroundService(i);
        else ctx.startService(i);
    }

    @Override
    public void onCreate() {
        super.onCreate();
        instance = this;
        createChannels();
        Notification n = runningNotification();
        if (Build.VERSION.SDK_INT >= 29) {
            startForeground(ID_RUNNING, n, ServiceInfo.FOREGROUND_SERVICE_TYPE_DATA_SYNC);
        } else {
            startForeground(ID_RUNNING, n);
        }
        WifiManager wm = (WifiManager) getApplicationContext().getSystemService(Context.WIFI_SERVICE);
        if (wm != null) {
            // Without this many phones drop the broadcast packets used to find devices.
            multicastLock = wm.createMulticastLock("alldrop-discovery");
            multicastLock.setReferenceCounted(false);
            multicastLock.acquire();
            wifiLock = wm.createWifiLock(WifiManager.WIFI_MODE_FULL_HIGH_PERF, "alldrop-transfer");
            wifiLock.setReferenceCounted(false);
        }
        PowerManager pm = (PowerManager) getSystemService(Context.POWER_SERVICE);
        if (pm != null) {
            cpuLock = pm.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "alldrop:transfer");
            cpuLock.setReferenceCounted(false);
        }
        netReceiver = new BroadcastReceiver() {
            @Override
            public void onReceive(Context c, Intent intent) {
                main.removeCallbacks(netCheck);
                main.postDelayed(netCheck, 2500);
            }
        };
        registerReceiver(netReceiver, new IntentFilter(ConnectivityManager.CONNECTIVITY_ACTION));
        Engine.start(this);
    }

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        if (intent != null && ACTION_STOP.equals(intent.getAction())) {
            stopping = true;
            Engine.stop();
            if (Build.VERSION.SDK_INT >= 24) stopForeground(STOP_FOREGROUND_REMOVE);
            else stopForeground(true);
            stopSelf();
            return START_NOT_STICKY;
        }
        Engine.start(this);
        return START_STICKY;
    }

    @Override
    public void onDestroy() {
        instance = null;
        main.removeCallbacks(netCheck);
        if (netReceiver != null) {
            try {
                unregisterReceiver(netReceiver);
            } catch (Exception ignored) {
            }
        }
        if (multicastLock != null && multicastLock.isHeld()) multicastLock.release();
        if (wifiLock != null && wifiLock.isHeld()) wifiLock.release();
        if (cpuLock != null && cpuLock.isHeld()) cpuLock.release();
        Engine.stop();
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    /** The engine process ended on its own: bring it back unless we are stopping. */
    static void engineExited(final Context ctx) {
        main.post(new Runnable() {
            @Override
            public void run() {
                EngineService s = instance;
                if (s == null || s.stopping) return;
                long now = System.currentTimeMillis();
                long wait = now - s.lastRestart < 10000 ? 5000 : 300;
                s.lastRestart = now;
                main.postDelayed(new Runnable() {
                    @Override
                    public void run() {
                        if (instance != null && !instance.stopping) Engine.start(instance);
                    }
                }, wait);
            }
        });
    }

    /** Keep CPU and Wi-Fi awake only while a transfer is running. */
    static void setBusy(final boolean busy) {
        main.post(new Runnable() {
            @Override
            public void run() {
                EngineService s = instance;
                if (s == null) return;
                try {
                    if (busy) {
                        if (s.cpuLock != null && !s.cpuLock.isHeld()) s.cpuLock.acquire(6 * 60 * 60 * 1000L);
                        if (s.wifiLock != null && !s.wifiLock.isHeld()) s.wifiLock.acquire();
                    } else {
                        if (s.cpuLock != null && s.cpuLock.isHeld()) s.cpuLock.release();
                        if (s.wifiLock != null && s.wifiLock.isHeld()) s.wifiLock.release();
                    }
                } catch (Exception ignored) {
                }
            }
        });
    }

    static void notifyOffer(Context ctx, String from, String count, String code) {
        NotificationManager nm = (NotificationManager) ctx.getSystemService(Context.NOTIFICATION_SERVICE);
        if (nm == null) return;
        Notification.Builder b = builder(ctx, CH_OFFERS)
                .setSmallIcon(R.drawable.ic_stat)
                .setContentTitle(ctx.getString(R.string.offer_title, from))
                .setContentText(code.length() == 6
                        ? ctx.getString(R.string.offer_code, count, code.substring(0, 3) + " " + code.substring(3))
                        : ctx.getString(R.string.offer_text, count))
                .setAutoCancel(true)
                .setContentIntent(openApp(ctx));
        if (Build.VERSION.SDK_INT < 26) {
            b.setPriority(Notification.PRIORITY_HIGH).setDefaults(Notification.DEFAULT_ALL);
        }
        if (Build.VERSION.SDK_INT >= 21) b.setCategory(Notification.CATEGORY_MESSAGE);
        nm.notify(ID_OFFER, b.build());
    }

    private static PendingIntent openApp(Context ctx) {
        Intent i = new Intent(ctx, MainActivity.class).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK | Intent.FLAG_ACTIVITY_SINGLE_TOP);
        int flags = PendingIntent.FLAG_UPDATE_CURRENT;
        if (Build.VERSION.SDK_INT >= 23) flags |= PendingIntent.FLAG_IMMUTABLE;
        return PendingIntent.getActivity(ctx, 0, i, flags);
    }

    @SuppressWarnings("deprecation")
    private static Notification.Builder builder(Context ctx, String channel) {
        if (Build.VERSION.SDK_INT >= 26) return new Notification.Builder(ctx, channel);
        return new Notification.Builder(ctx);
    }

    private Notification runningNotification() {
        Intent stop = new Intent(this, EngineService.class).setAction(ACTION_STOP);
        int flags = PendingIntent.FLAG_UPDATE_CURRENT;
        if (Build.VERSION.SDK_INT >= 23) flags |= PendingIntent.FLAG_IMMUTABLE;
        PendingIntent stopPi = PendingIntent.getService(this, 1, stop, flags);
        Notification.Builder b = builder(this, CH_SERVICE)
                .setSmallIcon(R.drawable.ic_stat)
                .setContentTitle(getString(R.string.running_title))
                .setContentText(getString(R.string.running_text))
                .setOngoing(true)
                .setContentIntent(openApp(this))
                .addAction(0, getString(R.string.stop), stopPi);
        if (Build.VERSION.SDK_INT < 26) b.setPriority(Notification.PRIORITY_LOW);
        return b.build();
    }

    private void createChannels() {
        if (Build.VERSION.SDK_INT < 26) return;
        NotificationManager nm = (NotificationManager) getSystemService(Context.NOTIFICATION_SERVICE);
        if (nm == null) return;
        NotificationChannel svc = new NotificationChannel(CH_SERVICE, getString(R.string.ch_service), NotificationManager.IMPORTANCE_LOW);
        svc.setShowBadge(false);
        nm.createNotificationChannel(svc);
        NotificationChannel offers = new NotificationChannel(CH_OFFERS, getString(R.string.ch_offers), NotificationManager.IMPORTANCE_HIGH);
        nm.createNotificationChannel(offers);
    }
}
