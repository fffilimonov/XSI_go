package main

import (
    "github.com/mattn/go-gtk/glib"
    "github.com/mattn/go-gtk/gdk"
    "github.com/mattn/go-gtk/gtk"
    "github.com/fffilimonov/OCIP_go"
    "os"
    "strings"
    "time"
)

func callWindow(ociConfig ocip.ConfigT) {
    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.SetTitle("Call Center")
    window.SetPosition(gtk.WIN_POS_CENTER)
    window.SetSizeRequest(300, 700)
    ocip.OCIPsend(ociConfig,"UserBasicCallLogsGetListRequest14sp4","userId=00070209393@spb.swisstok.ru","callLogType=Received")
//    ocip.OCIPsend(ociConfig,"00070209393@spb.swisstok.ru","UserBasicCallLogsGetListRequest14sp4","callLogType=Received")
    window.ShowAll()
}

func guiMain () {
    ch := make(chan CallInfo,100)
    var arg1 string
    arg1 = os.Args[1]
    if arg1 == "" {
        LogErr(nil,"no args")
        os.Exit (1)
    }
    var owner string = arg1
    var file string = "config"
    Config := ReadConfig(file)
    go clientMain(ch,Config)
    var ociConfig ocip.ConfigT
    ociConfig.Main.User=Config.Main.User
    ociConfig.Main.Password=Config.Main.Password
    ociConfig.Main.Host=Config.Main.Host
    ociConfig.Main.OCIPPort=Config.Main.OCIPPort

    timer := time.NewTimer(time.Second)
    timer.Stop()

    glib.ThreadInit(nil)
    gdk.ThreadsInit()
    gdk.ThreadsEnter()
    gtk.Init(&os.Args)

    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.SetTitle("Call Center")
    window.SetPosition(gtk.WIN_POS_CENTER)
    window.SetSizeRequest(700, 300)
    window.Connect("destroy", gtk.MainQuit)
    swin := gtk.NewScrolledWindow(nil, nil)
    swin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)

    b_av := gtk.NewButtonWithLabel("Available")
    b_av.Connect("clicked", func() {
            ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Available")
        })

    b_un := gtk.NewButtonWithLabel("Unavailable")
    b_un.Connect("clicked", func() {
            ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Unavailable")
        })

    b_wr := gtk.NewButtonWithLabel("Wrap-Up")
    b_wr.Connect("clicked", func() {
            ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Wrap-Up")
            timer.Reset(time.Second * Config.Main.Wraptime)
        })

    b_cl := gtk.NewButtonWithLabel("Calls")
    b_cl.Connect("clicked", func() {
            callWindow(ociConfig)
        })


    owner1 := gtk.NewLabel(owner)
    owner2 := gtk.NewLabel("")
    owner3 := gtk.NewLabel("")
    var count uint = 3

    var dlabel1,dlabel2,dlabel3 map[string]*gtk.Label
    dlabel1 = make(map[string]*gtk.Label)
    dlabel2 = make(map[string]*gtk.Label)
    dlabel3 = make(map[string]*gtk.Label)

    for _,target := range Config.Main.TargetID {
        if target != owner {
            count=count+1
            dlabel1[target] = gtk.NewLabel(target)
            dlabel2[target] = gtk.NewLabel("")
            dlabel3[target] = gtk.NewLabel("")
        }
    }

    table := gtk.NewTable(3, count, false)
    table.Attach(owner1,0,1,0,1,gtk.EXPAND,gtk.EXPAND,1,1)
    table.Attach(owner2,1,2,0,1,gtk.EXPAND,gtk.EXPAND,1,1)
    table.Attach(owner3,2,3,0,1,gtk.EXPAND,gtk.EXPAND,1,1)

    table.Attach(b_av,0,1,1,2,gtk.EXPAND,gtk.EXPAND,1,1)
    table.Attach(b_un,1,2,1,2,gtk.EXPAND,gtk.EXPAND,1,1)
    table.Attach(b_wr,2,3,1,2,gtk.EXPAND,gtk.EXPAND,1,1)

    var place uint = 1
    for _,target := range Config.Main.TargetID {
        if target != owner {
            place=place+1
            table.Attach(dlabel1[target],0,1,place,place+1,gtk.EXPAND,gtk.EXPAND,1,1)
            table.Attach(dlabel2[target],1,2,place,place+1,gtk.EXPAND,gtk.EXPAND,1,1)
            table.Attach(dlabel3[target],2,3,place,place+1,gtk.EXPAND,gtk.EXPAND,1,1)
        }
    }

    table.Attach(b_cl,1,2,count,count+1,gtk.EXPAND,gtk.EXPAND,1,1)

    swin.AddWithViewPort(table)
    window.Add(swin)
    window.SetDefaultSize(200, 200)
    window.ShowAll()

    go func() {
        for{
            select {
                case cinfo := <-ch:
                    if cinfo.Target==owner {
                        if cinfo.Pers == "Terminator" && cinfo.State == "Alerting" {
                            gdk.ThreadsEnter()
                            owner2.SetLabel(strings.Trim(cinfo.Addr,"tel:"))
                            gdk.ThreadsLeave()
                        }
                        if cinfo.Pers == "Terminator" && cinfo.State == "Released" {
                            gdk.ThreadsEnter()
                            owner2.SetLabel("")
                            gdk.ThreadsLeave()
                        }
                        if cinfo.CCstatus != "" {
                            gdk.ThreadsEnter()
                            owner3.SetLabel(cinfo.CCstatus)
                            gdk.ThreadsLeave()
                        }
                        if cinfo.CCstatuschanged != "" {
                            gdk.ThreadsEnter()
                            owner3.SetLabel(cinfo.CCstatuschanged)
                            gdk.ThreadsLeave()
                        }
                    } else {
                        if cinfo.Hook!="" {
                            gdk.ThreadsEnter()
                            tmp:=dlabel2[cinfo.Target]
                            tmp.SetLabel(cinfo.Hook)
                            dlabel2[cinfo.Target]=tmp
                            gdk.ThreadsLeave()
                        }
                        if cinfo.CCstatus != "" {
                            gdk.ThreadsEnter()
                            tmp:=dlabel3[cinfo.Target]
                            tmp.SetLabel(cinfo.CCstatus)
                            dlabel3[cinfo.Target]=tmp
                            gdk.ThreadsLeave()
                        }
                        if cinfo.CCstatuschanged != "" {
                            gdk.ThreadsEnter()
                            tmp:=dlabel3[cinfo.Target]
                            tmp.SetLabel(cinfo.CCstatuschanged)
                            dlabel3[cinfo.Target]=tmp
                            gdk.ThreadsLeave()
                        }
                    }
                case <-timer.C:
                    ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Available")
                default:
                    time.Sleep(time.Millisecond*10)
            }
        }
    }()

    gtk.Main()
}
