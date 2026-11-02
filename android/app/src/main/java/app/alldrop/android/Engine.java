package app.alldrop.android;

import android.content.Context;
import android.media.MediaScannerConnection;
import android.os.Build;
import android.os.Environment;
import android.provider.Settings;
import android.util.Log;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.net.Inet4Address;
import java.net.InetAddress;
import java.net.NetworkInterface;
import java.security.SecureRandom;
import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.Map;

/**
 * Runs the AllDrop engine (the same Go program used on Windows/macOS/Linux,
 * packaged as liballdrop.so) as a child process and relays its events.
 */
final class Engine {
    private static final String TAG = "AllDrop";

    static final String TOKEN = newToken();
    static volatile int port = 0;
    static volatile boolean failed = false;

    private static Process proc;
    private static String ipsKey = "";
    private static int generation = 0;

    private Engine() {}

    private static String newToken() {
        byte[] b = new byte[16];
        new SecureRandom().nextBytes(b);
        StringBuilder sb = new StringBuilder();
        for (byte x : b) sb.append(String.format("%02x", x & 0xff));
        return sb.toString();
    }

    static String localIPs() {
        List<String> out = new ArrayList<>();
        try {
            for (NetworkInterface ni : Collections.list(NetworkInterface.getNetworkInterfaces())) {
                if (!ni.isUp() || ni.isLoopback()) continue;
                for (InetAddress a : Collections.list(ni.getInetAddresses())) {
                    if (a instanceof Inet4Address && !a.isLoopbackAddress() && !a.isLinkLocalAddress()) {
                        out.add(a.getHostAddress());
                    }
                }
            }
        } catch (Exception e) {
            Log.w(TAG, "interfaces", e);
        }
        Collections.sort(out);
        StringBuilder sb = new StringBuilder();
        for (String s : out) {
            if (sb.length() > 0) sb.append(',');
            sb.append(s);
        }
        return sb.toString();
    }

    static String deviceName(Context ctx) {
        String n = null;
        try {
            if (Build.VERSION.SDK_INT >= 25) {
                n = Settings.Global.getString(ctx.getContentResolver(), Settings.Global.DEVICE_NAME);
            }
        } catch (Exception ignored) {
        }
        if (n == null || n.trim().isEmpty()) {
            String m = Build.MODEL == null ? "Android" : Build.MODEL;
            String b = Build.MANUFACTURER == null ? "" : Build.MANUFACTURER;
            n = m.toLowerCase().startsWith(b.toLowerCase()) ? m : (b.isEmpty() ? m : cap(b) + " " + m);
        }
        return n.trim();
    }

    private static String cap(String s) {
        return s.isEmpty() ? s : Character.toUpperCase(s.charAt(0)) + s.substring(1);
    }

    private static boolean canWrite(File dir) {
        if (dir == null) return false;
        try {
            if (!dir.isDirectory() && !dir.mkdirs()) return false;
            File probe = new File(dir, ".alldrop-probe");
            FileOutputStream o = new FileOutputStream(probe);
            o.write(1);
            o.close();
            probe.delete();
            return true;
        } catch (Exception e) {
            return false;
        }
    }

    /** Public Download/AllDrop when allowed, otherwise the app's own folder. */
    static File downloadDir(Context ctx) {
        File pub = new File(Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS), "AllDrop");
        if (canWrite(pub)) return pub;
        File ext = ctx.getExternalFilesDir(Environment.DIRECTORY_DOWNLOADS);
        File priv = ext != null ? new File(ext, "AllDrop") : new File(ctx.getFilesDir(), "Received");
        priv.mkdirs();
        return priv;
    }

    static synchronized boolean running() {
        if (proc == null) return false;
        try {
            proc.exitValue();
            return false;
        } catch (IllegalThreadStateException stillRunning) {
            return true;
        }
    }

    static synchronized void start(final Context app) {
        if (running()) return;
        final Context ctx = app.getApplicationContext();
        File bin = new File(ctx.getApplicationInfo().nativeLibraryDir, "liballdrop.so");
        File data = new File(ctx.getFilesDir(), "engine");
        data.mkdirs();
        File dl = downloadDir(ctx);
        ipsKey = localIPs();
        try {
            ProcessBuilder pb = new ProcessBuilder(bin.getAbsolutePath(),
                    "-android",
                    "-data", data.getAbsolutePath(),
                    "-downloads", dl.getAbsolutePath(),
                    "-name", deviceName(ctx),
                    "-ips", ipsKey);
            Map<String, String> env = pb.environment();
            env.put("HOME", ctx.getFilesDir().getAbsolutePath());
            env.put("TMPDIR", ctx.getCacheDir().getAbsolutePath());
            env.put("AD_OWNER_TOKEN", TOKEN);
            final Process p = pb.start();
            proc = p;
            failed = false;
            final int gen = ++generation;
            Thread out = new Thread(new Runnable() {
                @Override
                public void run() {
                    readEvents(ctx, p, gen);
                }
            }, "ad-events");
            out.setDaemon(true);
            out.start();
            Thread err = new Thread(new Runnable() {
                @Override
                public void run() {
                    drain(p.getErrorStream());
                }
            }, "ad-log");
            err.setDaemon(true);
            err.start();
            // stdin stays open: when this app dies the engine sees EOF and exits.
        } catch (Exception e) {
            Log.e(TAG, "engine start failed", e);
            failed = true;
            proc = null;
        }
    }

    static synchronized void stop() {
        generation++;
        port = 0;
        if (proc != null) {
            try {
                proc.getOutputStream().close();
            } catch (Exception ignored) {
            }
            proc.destroy();
            proc = null;
        }
    }

    static synchronized void restart(Context ctx) {
        stop();
        start(ctx);
    }

    /** Restart when the Wi-Fi network (and so our address) changed. */
    static synchronized void networkChanged(Context ctx) {
        if (!localIPs().equals(ipsKey)) {
            Log.i(TAG, "network changed, restarting engine");
            restart(ctx);
        }
    }

    private static void readEvents(Context ctx, Process p, int gen) {
        try {
            BufferedReader r = new BufferedReader(new InputStreamReader(p.getInputStream(), "UTF-8"));
            String line;
            while ((line = r.readLine()) != null) {
                if (!line.startsWith("ALLDROP\t")) continue;
                String[] f = line.split("\t");
                if (f.length < 3) continue;
                switch (f[1]) {
                    case "PORT":
                        port = Integer.parseInt(f[2].trim());
                        break;
                    case "SAVED":
                        MediaScannerConnection.scanFile(ctx, new String[]{f[2]}, null, null);
                        break;
                    case "OFFER":
                        EngineService.notifyOffer(ctx, f[2], f.length > 3 ? f[3] : "1", f.length > 4 ? f[4] : "");
                        break;
                    case "ACTIVE":
                        EngineService.setBusy("1".equals(f[2]));
                        break;
                    default:
                        break;
                }
            }
        } catch (Exception e) {
            Log.w(TAG, "event reader", e);
        }
        synchronized (Engine.class) {
            if (gen == generation) {
                port = 0;
                proc = null;
                EngineService.engineExited(ctx);
            }
        }
    }

    private static void drain(InputStream in) {
        try {
            BufferedReader r = new BufferedReader(new InputStreamReader(in, "UTF-8"));
            String line;
            while ((line = r.readLine()) != null) Log.i(TAG, line);
        } catch (Exception ignored) {
        }
    }
}
