package main

import (
    "github.com/google/gxui"
    "github.com/google/gxui/drivers/gl"
    "github.com/google/gxui/gxfont"
    "github.com/google/gxui/samples/flags"
    "net"
    "time"
)


func appMain(driver gxui.Driver) {
    guich := make(chan string,100)
    guiMain (driver,guich)
    go xsiMain(guich)
}

func guiMain (driver gxui.Driver, ch chan string) {
    theme := flags.CreateTheme(driver)
    font, err := driver.CreateFont(gxfont.Default, 25)
    if err != nil {
        panic(err)
    }
    window := theme.CreateWindow(400, 600, "Call")
    window.SetBackgroundBrush(gxui.CreateBrush(gxui.Gray50))
    label := theme.CreateLabel()
    label.SetFont(font)
    label.SetText("Starting...")
    window.AddChild(label)
    go func() {
        for{
            select {
                case str := <-ch:
                    driver.Call(func() {
                        label.SetText(str)
                    })
                default:
            }
        }
    }()
    window.OnClose(driver.Terminate)
}


func xsiMain (guich chan string) {
/* Read config */
    var file string = "config"
    Config := ReadConfig(file)
/* init default headers */
    def := MakeDef(Config)
/* chans */
    ch := make(chan string,100)
    datach := make(chan string,100)
    cCh := make(chan net.Conn)
/* start sybscription and reading to chan */
    go XSIresubscribe(Config,cCh)
    go XSIread(ch,cCh)
/* handle reading */
    go XSImain(Config,def,ch,datach)
    for {
        select {
            case data := <-datach:
                LogOut(data)
                cinfo := ParseData([]byte(data))
                Log2Out(cinfo.Target,cinfo.Pers,cinfo.State,cinfo.Addr,cinfo.Hook)
                if cinfo.Hook != "" {
                    Log2Out(cinfo.Target,cinfo.Hook)
                    guich<-ConcatStr("",cinfo.Target,"",cinfo.Hook)
                }
                if cinfo.Pers == "Terminator" && cinfo.State == "Alerting" {
                    Log2Out(cinfo.Target,"incoming",cinfo.Addr)
                    guich<-ConcatStr("",cinfo.Target," incomming ",cinfo.Addr)
                }
                if cinfo.Pers == "Terminator" && cinfo.State == "Released" {
                    Log2Out(cinfo.Target,"released",cinfo.Addr)
                    guich<-ConcatStr("",cinfo.Target," relesased ", cinfo.Addr)
                }
            default:
                time.Sleep(time.Millisecond*100)
        }
    }
}

func main() {
    gl.StartDriver(appMain)
}
