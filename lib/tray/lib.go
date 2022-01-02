package tray

import (
	_ "embed"
	"path"
	"runtime"
	"strings"

	"github.com/digitalcircle-com-br/devserver/lib/config"
	"github.com/digitalcircle-com-br/devserver/lib/server"
	"github.com/getlantern/systray"
	"github.com/skratchdot/open-golang/open"
)

//go:embed resources/logo_dc_square.png
var logo_png []byte

//go:embed resources/logo_dc_square.ico
var logo_ico []byte

func Run() {
	systray.Run(onReady, onExit)
}

func onReady() {
	if runtime.GOOS == "windows" {
		systray.SetIcon(logo_ico)
	} else {
		systray.SetIcon(logo_png)
	}

	systray.SetTitle("DC - DevServer")
	systray.SetTooltip("Digital Circle - Development Server & Gateway")

	mLogs := systray.AddMenuItem("Logs", "")
	mRestart := systray.AddMenuItem("Restart", "")
	mOpenDir := systray.AddMenuItem("Open Dir", "")
	mOpenConfig := systray.AddMenuItem("Open Config", "")
	mHelp := systray.AddMenuItem("Help", "")
	systray.AddSeparator()
	systray.AddMenuItem("Digital Circle® - V:0.0.13", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "")

	for {
		select {
		case <-mLogs.ClickedCh:
			if strings.HasPrefix(config.Cfg.Addr, ":") {
				open.Run("https://localhost" + config.Cfg.Addr + "/__log/index.html")
			} else {
				open.Run("https://" + config.Cfg.Addr + "/__log/index.html")
			}
		case <-mRestart.ClickedCh:
			server.StopHttpServer()
			server.StartHttpsServer()
		case <-mOpenDir.ClickedCh:
			open.Run(config.Wd)
		case <-mOpenConfig.ClickedCh:
			open.Run(path.Join(config.Wd, "config.yaml"))
		case <-mHelp.ClickedCh:
			open.Run("https://github.com/digitalcircle-com-br/devserver")
		case <-mQuit.ClickedCh:
			systray.Quit()
		}
	}

}

func onExit() {
	// clean up here
}
