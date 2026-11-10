package app.alldrop.android;

import android.Manifest;
import android.app.Activity;
import android.app.DownloadManager;
import android.content.ActivityNotFoundException;
import android.content.ClipData;
import android.content.Intent;
import android.content.pm.PackageManager;
import android.database.Cursor;
import android.graphics.Color;
import android.media.MediaScannerConnection;
import android.net.Uri;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.provider.DocumentsContract;
import android.provider.MediaStore;
import android.provider.OpenableColumns;
import android.view.Gravity;
import android.view.View;
import android.webkit.JavascriptInterface;
import android.webkit.MimeTypeMap;
import android.webkit.ValueCallback;
import android.webkit.WebChromeClient;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.FrameLayout;
import android.widget.TextView;
import android.widget.Toast;

import org.json.JSONArray;
import org.json.JSONObject;

import java.io.File;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.HttpURLConnection;
import java.net.URL;
import java.util.ArrayList;
import java.util.List;

/** Hosts the AllDrop web interface served by the local engine. */
public class MainActivity extends Activity {
    private static final int REQ_FILES = 1;
    private static final int REQ_PERMS = 2;
    private static final int REQ_PICK_FILES = 3;
    private static final int REQ_PICK_MEDIA = 4;
    private static final int REQ_PICK_TREE = 5;

    private final Handler handler = new Handler(Looper.getMainLooper());
    private WebView web;
    private TextView status;
    private ValueCallback<Uri[]> fileCallback;
    private int loadedPort = 0;
    private long waitStart;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        FrameLayout root = new FrameLayout(this);
        root.setBackgroundColor(Color.rgb(242, 244, 248));
        web = new WebView(this);
        web.setVisibility(View.INVISIBLE);
        root.addView(web, new FrameLayout.LayoutParams(-1, -1));
        status = new TextView(this);
        status.setText(R.string.starting);
        status.setTextColor(Color.rgb(90, 98, 110));
        status.setTextSize(16);
        status.setGravity(Gravity.CENTER);
        root.addView(status, new FrameLayout.LayoutParams(-1, -1));
        setContentView(root);
        if (Build.VERSION.SDK_INT >= 21) getWindow().setStatusBarColor(Color.rgb(10, 98, 216));

        WebSettings s = web.getSettings();
        s.setJavaScriptEnabled(true);
        s.setDomStorageEnabled(true);
        s.setAllowFileAccess(false);
        s.setAllowContentAccess(true);
        s.setMediaPlaybackRequiresUserGesture(true);
        web.addJavascriptInterface(new Bridge(), "AllDropNative");
        web.setWebViewClient(new WebViewClient() {
            @Override
            @SuppressWarnings("deprecation")
            public boolean shouldOverrideUrlLoading(WebView view, String url) {
                if (url.startsWith("http://127.0.0.1:")) return false;
                try {
                    startActivity(new Intent(Intent.ACTION_VIEW, Uri.parse(url)));
                } catch (Exception ignored) {
                }
                return true;
            }

            @Override
            public void onPageFinished(WebView view, String url) {
                web.setVisibility(View.VISIBLE);
                status.setVisibility(View.GONE);
            }
        });
        web.setWebChromeClient(new WebChromeClient() {
            @Override
            public boolean onShowFileChooser(WebView view, ValueCallback<Uri[]> cb, FileChooserParams params) {
                if (fileCallback != null) fileCallback.onReceiveValue(null);
                fileCallback = cb;
                Intent i = new Intent(Intent.ACTION_GET_CONTENT);
                i.addCategory(Intent.CATEGORY_OPENABLE);
                i.setType("*/*");
                i.putExtra(Intent.EXTRA_ALLOW_MULTIPLE, true);
                try {
                    startActivityForResult(Intent.createChooser(i, null), REQ_FILES);
                } catch (ActivityNotFoundException e) {
                    fileCallback = null;
                    return false;
                }
                return true;
            }
        });

        clearShareCache();
        askPermissions();
        EngineService.start(this);
        waitStart = System.currentTimeMillis();
        handler.post(waitForEngine);
        handleShare(getIntent());
    }

    private final Runnable waitForEngine = new Runnable() {
        @Override
        public void run() {
            int p = Engine.port;
            if (p > 0) {
                if (p != loadedPort) {
                    loadedPort = p;
                    web.loadUrl("http://127.0.0.1:" + p + "/?ot=" + Engine.TOKEN);
                }
            } else if (Engine.failed || System.currentTimeMillis() - waitStart > 20000) {
                status.setText(R.string.engine_failed);
                status.setVisibility(View.VISIBLE);
            }
            // keep watching: the engine restarts when the Wi-Fi network changes
            handler.postDelayed(this, p > 0 ? 1500 : 200);
        }
    };

    @Override
    protected void onNewIntent(Intent intent) {
        super.onNewIntent(intent);
        setIntent(intent);
        handleShare(intent);
    }

    @Override
    protected void onDestroy() {
        handler.removeCallbacks(waitForEngine);
        super.onDestroy();
    }

    @Override
    public void onBackPressed() {
        if (web.canGoBack()) web.goBack();
        else moveTaskToBack(true);
    }

    private void askPermissions() {
        List<String> need = new ArrayList<>();
        if (Build.VERSION.SDK_INT >= 33 && checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED) {
            need.add(Manifest.permission.POST_NOTIFICATIONS);
        }
        if (Build.VERSION.SDK_INT >= 23 && Build.VERSION.SDK_INT <= 29
                && checkSelfPermission(Manifest.permission.WRITE_EXTERNAL_STORAGE) != PackageManager.PERMISSION_GRANTED) {
            need.add(Manifest.permission.WRITE_EXTERNAL_STORAGE);
        }
        if (!need.isEmpty() && Build.VERSION.SDK_INT >= 23) {
            requestPermissions(need.toArray(new String[0]), REQ_PERMS);
        }
    }

    @Override
    public void onRequestPermissionsResult(int requestCode, String[] permissions, int[] results) {
        super.onRequestPermissionsResult(requestCode, permissions, results);
        for (int i = 0; i < permissions.length; i++) {
            if (Manifest.permission.WRITE_EXTERNAL_STORAGE.equals(permissions[i]) && results[i] == PackageManager.PERMISSION_GRANTED) {
                // switch the engine to the public Download folder
                Engine.restart(this);
            }
        }
    }

    @Override
    protected void onActivityResult(int requestCode, int resultCode, Intent data) {
        super.onActivityResult(requestCode, resultCode, data);
        if (requestCode == REQ_PICK_FILES || requestCode == REQ_PICK_MEDIA) {
            if (resultCode == RESULT_OK && data != null) {
                final List<Uri> uris = urisFrom(data);
                if (!uris.isEmpty()) {
                    Toast.makeText(this, R.string.preparing, Toast.LENGTH_SHORT).show();
                    new Thread(new Runnable() {
                        @Override
                        public void run() {
                            stageShared(uris, null);
                        }
                    }, "ad-pick").start();
                }
            }
            return;
        }
        if (requestCode == REQ_PICK_TREE) {
            if (resultCode == RESULT_OK && data != null && data.getData() != null) {
                final Uri tree = data.getData();
                Toast.makeText(this, R.string.preparing, Toast.LENGTH_SHORT).show();
                new Thread(new Runnable() {
                    @Override
                    public void run() {
                        stageTree(tree);
                    }
                }, "ad-pick-folder").start();
            }
            return;
        }
        if (requestCode != REQ_FILES || fileCallback == null) return;
        Uri[] result = null;
        if (resultCode == RESULT_OK && data != null) {
            ClipData clip = data.getClipData();
            if (clip != null && clip.getItemCount() > 0) {
                result = new Uri[clip.getItemCount()];
                for (int i = 0; i < clip.getItemCount(); i++) result[i] = clip.getItemAt(i).getUri();
            } else if (data.getData() != null) {
                result = new Uri[]{data.getData()};
            }
        }
        fileCallback.onReceiveValue(result);
        fileCallback = null;
    }

    // ---------- "Share" from other apps ----------

    private File shareDir() {
        return new File(getCacheDir(), "share");
    }

    private void clearShareCache() {
        if (Engine.running()) return; // engine may still be sending them
        deleteTree(shareDir());
    }

    private static void deleteTree(File f) {
        File[] kids = f.listFiles();
        if (kids != null) for (File k : kids) deleteTree(k);
        f.delete();
    }

    @SuppressWarnings("deprecation")
    private void handleShare(Intent intent) {
        if (intent == null) return;
        String action = intent.getAction();
        final List<Uri> uris = new ArrayList<>();
        String text = null;
        if (Intent.ACTION_SEND.equals(action)) {
            Uri u = intent.getParcelableExtra(Intent.EXTRA_STREAM);
            if (u != null) uris.add(u);
            else text = intent.getStringExtra(Intent.EXTRA_TEXT);
        } else if (Intent.ACTION_SEND_MULTIPLE.equals(action)) {
            ArrayList<Uri> list = intent.getParcelableArrayListExtra(Intent.EXTRA_STREAM);
            if (list != null) uris.addAll(list);
        } else {
            return;
        }
        intent.setAction(Intent.ACTION_MAIN); // handle once
        final String sharedText = text;
        if (uris.isEmpty() && sharedText == null) return;
        new Thread(new Runnable() {
            @Override
            public void run() {
                stageShared(uris, sharedText);
            }
        }, "ad-share").start();
    }

    private void stageShared(List<Uri> uris, String text) {
        File dir = new File(shareDir(), Long.toString(System.currentTimeMillis()));
        dir.mkdirs();
        JSONArray paths = new JSONArray();
        try {
            if (text != null) {
                File f = new File(dir, "AllDrop-text.txt");
                OutputStream o = new FileOutputStream(f);
                o.write(text.getBytes("UTF-8"));
                o.close();
                paths.put(f.getAbsolutePath());
            }
            int n = 0;
            for (Uri u : uris) {
                String name = displayName(u, ++n);
                File f = uniqueFile(dir, name);
                if (copyUri(u, f)) paths.put(f.getAbsolutePath());
            }
            postStage(paths, paths.length());
        } catch (Exception e) {
            android.util.Log.e("AllDrop", "share", e);
        }
    }

    private static List<Uri> urisFrom(Intent data) {
        List<Uri> out = new ArrayList<>();
        ClipData clip = data.getClipData();
        if (clip != null) {
            for (int i = 0; i < clip.getItemCount(); i++) {
                Uri u = clip.getItemAt(i).getUri();
                if (u != null) out.add(u);
            }
        } else if (data.getData() != null) {
            out.add(data.getData());
        }
        return out;
    }

    private boolean copyUri(Uri u, File dest) {
        try {
            InputStream in = getContentResolver().openInputStream(u);
            if (in == null) return false;
            OutputStream o = new FileOutputStream(dest);
            byte[] buf = new byte[1 << 20];
            int r;
            while ((r = in.read(buf)) > 0) o.write(buf, 0, r);
            o.close();
            in.close();
            return true;
        } catch (Exception e) {
            android.util.Log.w("AllDrop", "copy " + u, e);
            dest.delete();
            return false;
        }
    }

    /** A whole folder chosen with the system picker, sub folders included. */
    private void stageTree(Uri tree) {
        try {
            String rootId = DocumentsContract.getTreeDocumentId(tree);
            String rootName = queryName(DocumentsContract.buildDocumentUriUsingTree(tree, rootId));
            if (rootName == null || rootName.trim().isEmpty()) rootName = "Folder";
            File top = new File(new File(shareDir(), Long.toString(System.currentTimeMillis())), safeName(rootName));
            top.mkdirs();
            int count = copyTree(tree, rootId, top, 0);
            JSONArray paths = new JSONArray();
            paths.put(top.getAbsolutePath());
            postStage(paths, count);
        } catch (Exception e) {
            android.util.Log.e("AllDrop", "folder", e);
        }
    }

    private int copyTree(Uri tree, String docId, File dest, int depth) {
        if (depth > 32) return 0;
        int count = 0;
        Uri children = DocumentsContract.buildChildDocumentsUriUsingTree(tree, docId);
        Cursor c = null;
        try {
            c = getContentResolver().query(children, new String[]{
                    DocumentsContract.Document.COLUMN_DOCUMENT_ID,
                    DocumentsContract.Document.COLUMN_DISPLAY_NAME,
                    DocumentsContract.Document.COLUMN_MIME_TYPE}, null, null, null);
            if (c == null) return 0;
            while (c.moveToNext()) {
                String id = c.getString(0);
                String name = safeName(c.getString(1));
                String mime = c.getString(2);
                if (DocumentsContract.Document.MIME_TYPE_DIR.equals(mime)) {
                    File sub = new File(dest, name);
                    sub.mkdirs();
                    count += copyTree(tree, id, sub, depth + 1);
                } else if (copyUri(DocumentsContract.buildDocumentUriUsingTree(tree, id), uniqueFile(dest, name))) {
                    count++;
                }
            }
        } catch (Exception e) {
            android.util.Log.w("AllDrop", "tree", e);
        } finally {
            if (c != null) c.close();
        }
        return count;
    }

    private String queryName(Uri u) {
        try {
            Cursor c = getContentResolver().query(u, new String[]{OpenableColumns.DISPLAY_NAME}, null, null, null);
            if (c != null) {
                String n = c.moveToFirst() ? c.getString(0) : null;
                c.close();
                return n;
            }
        } catch (Exception ignored) {
        }
        return null;
    }

    private static String safeName(String n) {
        if (n == null || n.trim().isEmpty()) return "file";
        return n.replace('/', '_').replace('\\', '_');
    }

    /** Hands copied files/folders to the engine; they appear in the "To send" list. */
    private void postStage(JSONArray paths, final int count) {
        try {
            long deadline = System.currentTimeMillis() + 20000;
            while (Engine.port == 0 && System.currentTimeMillis() < deadline) Thread.sleep(200);
            JSONObject body = new JSONObject();
            body.put("paths", paths);
            body.put("temp", true);
            HttpURLConnection c = (HttpURLConnection) new URL("http://127.0.0.1:" + Engine.port + "/api/stage").openConnection();
            c.setRequestMethod("POST");
            c.setDoOutput(true);
            c.setRequestProperty("Content-Type", "application/json");
            c.setRequestProperty("X-AD", "1");
            c.setRequestProperty("X-AD-Owner", Engine.TOKEN);
            OutputStream os = c.getOutputStream();
            os.write(body.toString().getBytes("UTF-8"));
            os.close();
            c.getResponseCode();
            c.disconnect();
            runOnUiThread(new Runnable() {
                @Override
                public void run() {
                    Toast.makeText(MainActivity.this, getString(R.string.shared_ready, count), Toast.LENGTH_LONG).show();
                }
            });
        } catch (Exception e) {
            android.util.Log.e("AllDrop", "stage", e);
        }
    }

    private String displayName(Uri u, int n) {
        String name = null;
        try {
            Cursor c = getContentResolver().query(u, new String[]{OpenableColumns.DISPLAY_NAME}, null, null, null);
            if (c != null) {
                if (c.moveToFirst()) name = c.getString(0);
                c.close();
            }
        } catch (Exception ignored) {
        }
        if (name == null || name.trim().isEmpty()) {
            name = u.getLastPathSegment();
            String ext = MimeTypeMap.getSingleton().getExtensionFromMimeType(getContentResolver().getType(u));
            if (name == null || name.isEmpty()) name = "file-" + n;
            if (ext != null && !name.contains(".")) name += "." + ext;
        }
        return name.replace('/', '_').replace('\\', '_');
    }

    private static File uniqueFile(File dir, String name) {
        File f = new File(dir, name);
        int i = 2;
        while (f.exists()) {
            int dot = name.lastIndexOf('.');
            String base = dot > 0 ? name.substring(0, dot) : name;
            String ext = dot > 0 ? name.substring(dot) : "";
            f = new File(dir, base + " (" + (i++) + ")" + ext);
        }
        return f;
    }

    // ---------- bridge used by the web UI ----------

    private void startPicker(String kind) {
        try {
            if ("folder".equals(kind)) {
                startActivityForResult(new Intent(Intent.ACTION_OPEN_DOCUMENT_TREE), REQ_PICK_TREE);
            } else if ("media".equals(kind)) {
                Intent i;
                if (Build.VERSION.SDK_INT >= 33) {
                    i = new Intent(MediaStore.ACTION_PICK_IMAGES);
                    i.putExtra(MediaStore.EXTRA_PICK_IMAGES_MAX, MediaStore.getPickImagesMaxLimit());
                } else {
                    i = new Intent(Intent.ACTION_GET_CONTENT);
                    i.addCategory(Intent.CATEGORY_OPENABLE);
                    i.setType("*/*");
                    i.putExtra(Intent.EXTRA_MIME_TYPES, new String[]{"image/*", "video/*"});
                    i.putExtra(Intent.EXTRA_ALLOW_MULTIPLE, true);
                }
                startActivityForResult(i, REQ_PICK_MEDIA);
            } else {
                Intent i = new Intent(Intent.ACTION_OPEN_DOCUMENT);
                i.addCategory(Intent.CATEGORY_OPENABLE);
                i.setType("*/*");
                i.putExtra(Intent.EXTRA_ALLOW_MULTIPLE, true);
                startActivityForResult(i, REQ_PICK_FILES);
            }
        } catch (ActivityNotFoundException e) {
            Toast.makeText(this, R.string.no_viewer, Toast.LENGTH_SHORT).show();
        }
    }

    private final class Bridge {
        /** Native pickers: "files" (several), "folder" (whole folder) or "media" (photos and videos). */
        @JavascriptInterface
        public void pick(final String kind) {
            runOnUiThread(new Runnable() {
                @Override
                public void run() {
                    startPicker(kind);
                }
            });
        }

        @JavascriptInterface
        public void openDownloads() {
            runOnUiThread(new Runnable() {
                @Override
                public void run() {
                    try {
                        startActivity(new Intent(DownloadManager.ACTION_VIEW_DOWNLOADS).addFlags(Intent.FLAG_ACTIVITY_NEW_TASK));
                    } catch (Exception e) {
                        Toast.makeText(MainActivity.this, R.string.no_viewer, Toast.LENGTH_SHORT).show();
                    }
                }
            });
        }

        @JavascriptInterface
        public void openFile(final String path) {
            String ext = MimeTypeMap.getFileExtensionFromUrl(Uri.fromFile(new File(path)).toString());
            final String mime = ext == null ? null : MimeTypeMap.getSingleton().getMimeTypeFromExtension(ext.toLowerCase());
            MediaScannerConnection.scanFile(MainActivity.this, new String[]{path}, null, new MediaScannerConnection.OnScanCompletedListener() {
                @Override
                public void onScanCompleted(String p, final Uri uri) {
                    runOnUiThread(new Runnable() {
                        @Override
                        public void run() {
                            Uri target = uri;
                            if (target == null && Build.VERSION.SDK_INT < 24) target = Uri.fromFile(new File(path));
                            if (target == null) {
                                Toast.makeText(MainActivity.this, R.string.no_viewer, Toast.LENGTH_SHORT).show();
                                return;
                            }
                            Intent i = new Intent(Intent.ACTION_VIEW);
                            i.setDataAndType(target, mime != null ? mime : "*/*");
                            i.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION | Intent.FLAG_ACTIVITY_NEW_TASK);
                            try {
                                startActivity(i);
                            } catch (Exception e) {
                                Toast.makeText(MainActivity.this, R.string.no_viewer, Toast.LENGTH_SHORT).show();
                            }
                        }
                    });
                }
            });
        }
    }
}
