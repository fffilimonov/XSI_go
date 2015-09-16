package main

import (
    "github.com/mattn/go-gtk/glib"
    "github.com/mattn/go-gtk/gdk"
    "github.com/mattn/go-gtk/gtk"
    "github.com/fffilimonov/OCIP_go"
    "strings"
    "time"
)

func callWindow(ociConfig ocip.ConfigT) {
    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.SetTitle("Call Center")
    window.SetPosition(gtk.WIN_POS_CENTER)
    window.SetSizeRequest(400, 800)
    ocip.OCIPsend(ociConfig,"UserBasicCallLogsGetListRequest14sp4","userId=00070209393@spb.swisstok.ru","callLogType=Received")
    window.ShowAll()
}

func guiMain (confglobal string,conflocal string) {
    ch := make(chan CallInfo,100)
    Config := ReadConfig(confglobal)
    Configlocal := ReadConfiglocal(conflocal)
    owner:=Configlocal.Main.Owner
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
    gtk.Init(nil)
//icons to pixbuf
    im_call := gtk.NewImageFromFile("./Call-Ringing-48.png")
    pix_call := im_call.GetPixbuf()
    im_blank := gtk.NewImageFromFile("./Empty-48.png")
    pix_blank := im_blank.GetPixbuf()
    im_green := gtk.NewImageFromFile("./Green-ball-48.png")
    pix_green := im_green.GetPixbuf()
    im_grey := gtk.NewImageFromFile("./Grey-ball-48.png")
    pix_grey := im_grey.GetPixbuf()
    im_yellow := gtk.NewImageFromFile("./Yellow-ball-48.png")
    pix_yellow := im_yellow.GetPixbuf()

    window := gtk.NewWindow(gtk.WINDOW_TOPLEVEL)
    window.SetTitle("Call Center")
    window.SetIcon(pix_call)
    window.SetPosition(gtk.WIN_POS_CENTER)
    window.SetSizeRequest(700, 700)
    window.Connect("destroy", gtk.MainQuit)
    swin := gtk.NewScrolledWindow(nil, nil)
    swin.SetPolicy(gtk.POLICY_AUTOMATIC, gtk.POLICY_AUTOMATIC)

    b_av := gtk.NewButtonWithLabel("Доступен")
    b_av.SetCanFocus(false)
    b_av.Connect("clicked", func() {
            ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Available")
        })

    b_un := gtk.NewButtonWithLabel("Недоступен")
    b_un.SetCanFocus(false)
    b_un.Connect("clicked", func() {
            ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Unavailable")
        })

    b_wr := gtk.NewButtonWithLabel("Обработка")
    b_wr.SetCanFocus(false)
    b_wr.Connect("clicked", func() {
            ocip.OCIPsend(ociConfig,"UserCallCenterModifyRequest19",ConcatStr("","userId=",owner),"agentACDState=Wrap-Up")
        })

    b_cl := gtk.NewButtonWithLabel("Calls")
    b_cl.SetCanFocus(false)
    b_cl.Connect("clicked", func() {
            callWindow(ociConfig)
        })

    names := make(map[string]string)
    for iter,target := range Config.Main.TargetID {
        names[target]=Config.Main.Name[iter]
    }

    owner1 := gtk.NewLabel(names[owner])
    owner2 := gtk.NewLabel("")
    owner3 := gtk.NewImage()

    var count uint = 3

    dlabel1 := make(map[string]*gtk.Label)
    dlabel2 := make(map[string]*gtk.Image)
    dlabel3 := make(map[string]*gtk.Image)

    for _,target := range Config.Main.TargetID {
        if target != owner {
            count=count+1
            dlabel1[target] = gtk.NewLabel(names[target])
            dlabel2[target] = gtk.NewImage()
            dlabel3[target] = gtk.NewImage()
        }
    }

    table := gtk.NewTable(3, count, false)
    table.Attach(owner1,0,1,0,1,gtk.EXPAND,gtk.EXPAND,1,1)
    table.Attach(owner3,1,2,0,1,gtk.EXPAND,gtk.EXPAND,1,1)
    table.Attach(owner2,2,3,0,1,gtk.EXPAND,gtk.EXPAND,1,1)

    table.Attach(b_av,0,1,1,2,gtk.EXPAND,gtk.EXPAND,1,1)
    table.Attach(b_un,1,2,1,2,gtk.EXPAND,gtk.EXPAND,1,1)
    table.Attach(b_wr,2,3,1,2,gtk.EXPAND,gtk.EXPAND,1,1)

    var place uint = 1
    for _,target := range Config.Main.TargetID {
        if target != owner {
            place=place+1
            table.Attach(dlabel1[target],0,1,place,place+1,gtk.EXPAND,gtk.EXPAND,1,1)
            table.Attach(dlabel3[target],1,2,place,place+1,gtk.EXPAND,gtk.EXPAND,1,1)
            table.Attach(dlabel2[target],2,3,place,place+1,gtk.EXPAND,gtk.EXPAND,1,1)
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
                        if cinfo.CCstatus != "" || cinfo.CCstatuschanged != "" {
                            var ccstatus string
                            if cinfo.CCstatus != "" {
                                ccstatus=cinfo.CCstatus
                            }
                            if cinfo.CCstatuschanged != "" {
                                ccstatus=cinfo.CCstatuschanged
                            }
                            gdk.ThreadsEnter()
                                if ccstatus == "Available" {
                                    owner3.SetFromPixbuf(pix_green)
                                } else if ccstatus == "Wrap-Up" {
                                    owner3.SetFromPixbuf(pix_yellow)
                                    timer.Reset(time.Second * Config.Main.Wraptime)
                                }else{
                                    owner3.SetFromPixbuf(pix_grey)
                                }
                            gdk.ThreadsLeave()
                        }
                    } else {
                        if cinfo.Hook!="" {
                            gdk.ThreadsEnter()
                            tmp:=dlabel2[cinfo.Target]
                            if cinfo.Hook == "Off-Hook" {
                                tmp.SetFromPixbuf(pix_call)
                            } else {
                                tmp.SetFromPixbuf(pix_blank)
                            }
                            dlabel2[cinfo.Target]=tmp
                            gdk.ThreadsLeave()
                        }
                        if cinfo.CCstatus != "" || cinfo.CCstatuschanged != "" {
                            var ccstatus string
                            if cinfo.CCstatus != "" {
                                ccstatus=cinfo.CCstatus
                            }
                            if cinfo.CCstatuschanged != "" {
                                ccstatus=cinfo.CCstatuschanged
                            }
                            gdk.ThreadsEnter()
                            tmp:=dlabel3[cinfo.Target]
                            if ccstatus == "Available" {
                                tmp.SetFromPixbuf(pix_green)
                            } else if ccstatus == "Wrap-Up" {
                                tmp.SetFromPixbuf(pix_yellow)
                            }else{
                                tmp.SetFromPixbuf(pix_grey)
                            }
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
