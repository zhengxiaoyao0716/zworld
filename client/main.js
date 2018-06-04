const { app, BrowserWindow } = require('electron')

let win

function createWindow() {
  // Create the browser window.
  win = new BrowserWindow({
    width: 800, height: 600,
    frame: true, fullscreen: false, resizable: true,
    webPreferences: {
      nodeIntegration: false,
    },
  })
  win.setMenuBarVisibility(false)
  win.maximize();

  // win.loadURL(url.format({
  //   pathname: path.join(__dirname, './../browser/index.html'),
  //   protocol: 'file:',
  //   slashes: true
  // }))
  win.loadURL('http://localhost:2017')

  // 打开开发者工具
  win.webContents.openDevTools()

  win.on('closed', () => {
    win = null
  })
}

app.on('ready', createWindow)

app.on('window-all-closed', () => {
  // 在 macOS 上，除非用户用 Cmd + Q 确定地退出，
  // 否则绝大部分应用及其菜单栏会保持激活。
  if (process.platform !== 'darwin') {
    app.quit()
  }
})

app.on('activate', () => {
  // 在macOS上，当单击dock图标并且没有其他窗口打开时，
  // 通常在应用程序中重新创建一个窗口。
  if (win === null) {
    createWindow()
  }
})
