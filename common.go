package main

import (
    "encoding/base64"
    "fmt"
    "gopkg.in/gcfg.v1"
    "math/rand"
    "os"
    "strings"
    "time"
)

type ConfigT struct {
    Main struct {
        User string
        Password string
        Host string
        HTTPHost string
        Port string
        Expires string
        Target string
        TargetID string
        Event string
    }
}

func ReadConfig(Configfile string) ConfigT {
    var Config ConfigT
    err := gcfg.ReadFileInto(&Config, Configfile)
    if err != nil {
        LogErr(err,"Config file is missing:", Configfile)
        os.Exit (1)
    }
    return Config
}

func ConcatStr(sep string, args ... string) string {
    return strings.Join(args, sep)
}

func MakeAuth(User string, Password string) string {
    var concatedstr string = ConcatStr(":",User,Password)
    return base64.StdEncoding.EncodeToString([]byte(concatedstr))
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
func randSeq(n int) string {
    rand.Seed(time.Now().UnixNano())
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func LogErr (err error,args ... string) {
    fmt.Fprint(os.Stderr,time.Now(),args,err,"\n")
}

func LogOut (log string) {
    fmt.Fprint(os.Stdout,log,"\n\n")
}

func Log2Out (args ... string) {
    fmt.Fprint(os.Stdout,args,"\n\n")
}
