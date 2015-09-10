package main

import (
    "github.com/google/gxui"
    "github.com/google/gxui/gxfont"
    "github.com/google/gxui/samples/flags"
    "os"
    "strings"
    "time"
)

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

func setButtons(button1 gxui.Button, button2 gxui.Button, button3 gxui.Button, status string) {
    if status == "Available" {
        button1.SetChecked(true)
        button2.SetChecked(false)
        button3.SetChecked(false)
    }
    if status == "Unavailable" {
        button1.SetChecked(false)
        button2.SetChecked(true)
        button3.SetChecked(false)
    }
    if status == "Wrap-Up" {
        button1.SetChecked(false)
        button2.SetChecked(false)
        button3.SetChecked(true)
    }
}


//    button1.SetBackgroundBrush(gxui.CreateBrush(gxui.Red80))
//    button1.SetBorderPen(gxui.CreatePen(5, gxui.Gray80))
func initButton(theme gxui.Theme,Config ConfigT,owner string,status string) gxui.Button {
    button := theme.CreateButton()
    button.SetText(status)
    button.SetType(gxui.PushButton)
    button.SetChecked(false)
    button.OnClick(func(gxui.MouseEvent){
        OCIPsend(Config,owner,status)
    })
    return button
}

func guiMain (driver gxui.Driver, ch chan CallInfo, owner string, Config ConfigT) {
    theme := flags.CreateTheme(driver)
    font, _ := driver.CreateFont(gxfont.Default, 15)
    font1, _ := driver.CreateFont(gxfont.Default, 35)
    window := theme.CreateWindow(600, 300, "Call")
    window.SetBackgroundBrush(gxui.CreateBrush(gxui.White))
    layout := theme.CreateLinearLayout()
    layout.SetBackgroundBrush(gxui.CreateBrush(gxui.White))
    layout.SetDirection(gxui.TopToBottom)


//buttons
    button1 := initButton(theme,Config,owner,"Available")
    button2 := initButton(theme,Config,owner,"Unavailable")
    button3 := initButton(theme,Config,owner,"Wrap-Up")
//button

//owner
    label1 := theme.CreateLabel()
    label1.SetColor(gxui.Black)
    label1.SetFont(font1)
    label1.SetHorizontalAlignment(gxui.AlignCenter)
    label1.SetText("")
    layout.AddChild(label1)

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
//set owners table
    table.SetChildAt(0, 0, 1, 1, button1)
    table.SetChildAt(1, 0, 1, 1, button2)
    table.SetChildAt(2, 0, 1, 1, button3)
    count=0
    for _,target := range Config.Main.TargetID {
        if target != owner {
            count=count+1
            table.SetChildAt(0, count, 1, 1, dcell1[target])
            table.SetChildAt(1, count, 1, 1, dcell2[target])
            table.SetChildAt(2, count, 1, 1, dcell3[target])
            var inittarget gxui.Label = dlabel1[target]
            inittarget.SetText(target)
            dlabel1[target] = inittarget
        }
    }
layout.AddChild(table)
    window.AddChild(layout)

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
                            if cinfo.CCstatus != "" {
                                setButtons(button1,button2,button3,cinfo.CCstatus)
                            }
                            if cinfo.CCstatuschanged != "" {
                                setButtons(button1,button2,button3,cinfo.CCstatuschanged)
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
