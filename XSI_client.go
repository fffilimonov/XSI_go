package main

import (
    "net"
    "os"
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
                cinfo := ParseData([]byte(data))
                guich<-cinfo
            default:
                time.Sleep(time.Millisecond*10)
        }
    }
}

func main() {
    larg:=len(os.Args)
    if larg < 3 {
        LogErr(nil,"no args")
        os.Exit (1)
    }
    var globalconf string = os.Args[1]
    var localconf string = os.Args[2]
    guiMain(globalconf,localconf)
}
