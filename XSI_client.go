package main

import (
    "github.com/google/gxui"
    "github.com/google/gxui/drivers/gl"
    "github.com/google/gxui/gxfont"
    "github.com/google/gxui/samples/flags"
    "net"
    "os"
    "time"
    "strings"
)

func guiinit(theme gxui.Theme,font gxui.Font) (gxui.Label,gxui.LinearLayout) {
    var label gxui.Label
    var cell gxui.LinearLayout
    label = theme.CreateLabel()
    label.SetColor(gxui.Black)
    label.SetFont(font)
    cell = theme.CreateLinearLayout()
    cell.SetBackgroundBrush(gxui.CreateBrush(gxui.White))
    cell.SetHorizontalAlignment(gxui.AlignLeft)
    cell.AddChild(label)
    return label,cell
}

func appMain(driver gxui.Driver) {
    guich := make(chan CallInfo,100)
    var arg1 string
    arg1 = os.Args[1]
    if arg1 == "" {
        LogErr(nil,"no args")
        os.Exit (1)
    }
    var file string = "config"
    Config := ReadConfig(file)
    go clientMain(guich,Config)
    guiMain (driver,guich,arg1,Config)
}

func guiMain (driver gxui.Driver, ch chan CallInfo, owner string, Config ConfigT) {
    theme := flags.CreateTheme(driver)
    font, err := driver.CreateFont(gxfont.Default, 15)
    if err != nil {
        panic(err)
    }
    window := theme.CreateWindow(600, 300, "Call")
    window.SetBackgroundBrush(gxui.CreateBrush(gxui.White))


//button
    layout := theme.CreateLinearLayout()
    layout.SetBackgroundBrush(gxui.CreateBrush(gxui.White))
    layout.SetDirection(gxui.LeftToRight)

    button := func(name string, action func()) gxui.Button {
        b := theme.CreateButton()
        b.SetText(name)
        b.OnClick(func(gxui.MouseEvent){action()})
        layout.AddChild(b)
        return b
    }

    button("Available",func() {
        Log2Out("pressed Avail")
        OCIPsend(Config,owner,"Available")
    },)
    button("Unavailable",func() {
        Log2Out("pressed Unavail")
        OCIPsend(Config,owner,"Unavailable")
    },)
    window.AddChild(layout)
//button

//for incoming call
    label1 := theme.CreateLabel()
    label1.SetColor(gxui.Black)
    label1.SetFont(font)
    label1.SetText("")

    cell1 := theme.CreateLinearLayout()
    cell1.SetBackgroundBrush(gxui.CreateBrush(gxui.White))
    cell1.SetHorizontalAlignment(gxui.AlignLeft)
    cell1.AddChild(label1)

//targets
    var dlabel1,dlabel2,dlabel3 map[string]gxui.Label
    var dcell1,dcell2,dcell3 map[string]gxui.LinearLayout
    count:=1
    dlabel1 = make(map[string]gxui.Label)
    dcell1 = make(map[string]gxui.LinearLayout)
    dlabel2 = make(map[string]gxui.Label)
    dcell2 = make(map[string]gxui.LinearLayout)
    dlabel3 = make(map[string]gxui.Label)
    dcell3 = make(map[string]gxui.LinearLayout)

     for _,target := range Config.Main.TargetID {
        count=count+1
        dlabel1[target],dcell1[target] = guiinit(theme,font)
        dlabel2[target],dcell2[target] = guiinit(theme,font)
        dlabel3[target],dcell3[target] = guiinit(theme,font)
    }

    table := theme.CreateTableLayout()
    table.SetGrid(3, count) // rows, columns

    // row, column, horizontal span, vertical span
    table.SetChildAt(1, 0, 1, 1, cell1)
    count=0
    for _,target := range Config.Main.TargetID {
        count=count+1
        table.SetChildAt(0, count, 1, 1, dcell1[target])
        table.SetChildAt(1, count, 1, 1, dcell2[target])
        table.SetChildAt(2, count, 1, 1, dcell3[target])
        var inittarget gxui.Label = dlabel1[target]
        inittarget.SetText(target)
        dlabel1[target] = inittarget
    }

    window.AddChild(table)

    go func() {
        for{
            select {
                case cinfo := <-ch:
                    driver.Call(func() {
                        if cinfo.Target==owner {
                            if cinfo.Pers == "Terminator" && cinfo.State == "Alerting" {
                                label1.SetText(strings.Trim(cinfo.Addr,"tel:"))
                            }
                            if cinfo.Pers == "Terminator" && cinfo.State == "Released" {
                                label1.SetText("")
                            }
                        }
                        var tmpset2 gxui.Label = dlabel2[cinfo.Target]
                        var tmpset3 gxui.Label = dlabel3[cinfo.Target]
                        if cinfo.Hook!="" {
                            tmpset2.SetText(cinfo.Hook)
                        }
                        if cinfo.CCstatus != "" {
                            tmpset3.SetText(cinfo.CCstatus)
                        }
                        if cinfo.CCstatuschanged != "" {
                            tmpset3.SetText(cinfo.CCstatuschanged)
                        }
                        dlabel2[cinfo.Target] = tmpset2
                        dlabel3[cinfo.Target] = tmpset3
                    })
                default:
                    time.Sleep(time.Millisecond*10)
            }
        }
    }()
    window.OnClose(driver.Terminate)
}


func clientMain (guich chan CallInfo,Config ConfigT) {
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
                Log2Out(cinfo.Target,cinfo.Pers,cinfo.State,cinfo.Addr,cinfo.Hook,cinfo.CCstatus,cinfo.CCstatuschanged)
                guich<-cinfo
            default:
                time.Sleep(time.Millisecond*10)
        }
    }
}

func main() {
    gl.StartDriver(appMain)
}
