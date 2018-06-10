const { app, BrowserWindow, ipcMain } = require('electron');

let win;

const startUp = () => {
  const startUp = new BrowserWindow({ width: 600, height: 200, frame: false, transparent: true });
  startUp.loadFile('./login.html');
  win = startUp;
};
app.on('ready', startUp);

const login = (_event, { url, dev, full }) => {
  const mainWin = new BrowserWindow({
    width: 800, height: 600,
    frame: true, fullscreen: false, resizable: true,
    webPreferences: {
      nodeIntegration: false,
    },
  });
  mainWin.loadURL(url);
  dev && mainWin.webContents.openDevTools();
  full ? mainWin.setFullScreen(true) : mainWin.maximize();

  mainWin.setMenuBarVisibility(false)
  win.close(); // close startUp.
  win = mainWin;
};
ipcMain.once('login', login);


app.on('window-all-closed', () => {
  // 在 macOS 上，除非用户用 Cmd + Q 确定地退出，
  // 否则绝大部分应用及其菜单栏会保持激活。
  process.platform === 'darwin' || app.quit();
});

app.on('activate', () => {
  // 在macOS上，当单击dock图标并且没有其他窗口打开时，
  // 通常在应用程序中重新创建一个窗口。
  win === null && createWindow();
});
