package main

import (
    "net"
    "time"
)

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
    guiMain()
}
