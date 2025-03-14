# WinRM for Go

_Note_: if you're looking for the `winrm` command-line tool, this has been splitted from this project and is available at [winrm-cli](https://github.com/marslin2023/winrm-go-cli)

This is a Go library to execute remote commands on Windows machines through
the use of WinRM/WinRS.

_Note_: this library doesn't support domain users (it doesn't support GSSAPI nor Kerberos). It's primary target is to execute remote commands on EC2 windows machines.

[![Build Status](https://travis-ci.org/masterzen/winrm.svg?branch=master)](https://travis-ci.org/masterzen/winrm)
[![Coverage Status](https://coveralls.io/repos/masterzen/winrm/badge.png)](https://coveralls.io/r/masterzen/winrm)

## Contact

- Bugs: https://github.com/marslin2023/winrm-go/issues


## Getting Started
WinRM is available on Windows Server 2008 and up. This project natively supports basic authentication for local accounts, see the steps in the next section on how to prepare the remote Windows machine for this scenario. The authentication model is pluggable, see below for an example on using Negotiate/NTLM authentication (e.g. for connecting to vanilla Azure VMs) or Kerberos authentication (using domain accounts).

_Note_: This library only supports Golang 1.7+

### Preparing the remote Windows machine for Basic authentication
This project supports only basic authentication for local accounts (domain users are not supported). The remote windows system must be prepared for winrm:

_For a PowerShell script to do what is described below in one go, check [Richard Downer's blog](http://www.frontiertown.co.uk/2011/12/overthere-control-windows-from-java/)_

On the remote host, a PowerShell prompt, using the __Run as Administrator__ option and paste in the following lines:

		winrm quickconfig
		y
		winrm set winrm/config/service/Auth '@{Basic="true"}'
		winrm set winrm/config/service '@{AllowUnencrypted="true"}'
		winrm set winrm/config/winrs '@{MaxMemoryPerShellMB="1024"}'

__N.B.:__ The Windows Firewall needs to be running to run this command. See [Microsoft Knowledge Base article #2004640](http://support.microsoft.com/kb/2004640).

__N.B.:__ Do not disable Negotiate authentication as the `winrm` command itself uses this for internal authentication, and you risk getting a system where `winrm` doesn't work anymore.

__N.B.:__ The `MaxMemoryPerShellMB` option has no effects on some Windows 2008R2 systems because of a WinRM bug. Make sure to install the hotfix described [Microsoft Knowledge Base article #2842230](http://support.microsoft.com/kb/2842230) if you need to run commands that use more than 150MB of memory.

For more information on WinRM, please refer to <a href="http://msdn.microsoft.com/en-us/library/windows/desktop/aa384426(v=vs.85).aspx">the online documentation at Microsoft's DevCenter</a>.

### Preparing the remote Windows machine for kerberos authentication
This project supports domain users via kerberos authentication. The remote windows system must be prepared for winrm:

On the remote host, a PowerShell prompt, using the __Run as Administrator__ option and paste in the following lines:

                winrm quickconfig
                y
                winrm set winrm/config/service '@{AllowUnencrypted="true"}'
                winrm set winrm/config/winrs '@{MaxMemoryPerShellMB="1024"}'

All __N.B__ points of "Preparing the remote Windows machine for Basic authentication" also applies.


### Building the winrm go and executable

You can build winrm from source:

```sh
git clone https://github.com/marslin2023/winrm-go
cd winrm
make
```

_Note_: this winrm code doesn't depend anymore on [Gokogiri](https://github.com/moovweb/gokogiri) which means it is now in pure Go.

_Note_: you need go 1.5+. Please check your installation with

```
go version
```

## Command-line usage

For command-line usage check the [winrm-cli project](https://github.com/marslin2023/winrm-go-cli)

## Library Usage

**Warning the API might be subject to change.**

For the fast version (this doesn't allow to send input to the command) and it's using HTTP as the transport:

```go
package main

import (
	"github.com/marslin2023/winrm-go"
	"os"
)

endpoint := winrm.NewEndpoint(host, 5986, false, false, nil, nil, nil, 0)
client, err := winrm.NewClient(endpoint, "Administrator", "secret")
if err != nil {
	panic(err)
}
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
client.RunWithContext(ctx, "ipconfig /all", os.Stdout, os.Stderr)
```

or
```go
package main
import (
  "github.com/marslin2023/winrm-go"
  "fmt"
  "os"
)

endpoint := winrm.NewEndpoint("localhost", 5985, false, false, nil, nil, nil, 0)
client, err := winrm.NewClient(endpoint,"Administrator", "secret")
if err != nil {
	panic(err)
}

ctx, cancel := context.WithCancel(context.Background())
defer cancel()
_, err := client.RunWithContextWithInput(ctx, "ipconfig", os.Stdout, os.Stderr, os.Stdin)
if err != nil {
	panic(err)
}

```

By passing a TransportDecorator in the Parameters struct it is possible to use different Transports (e.g. NTLM)

```go
package main
import (
  "github.com/marslin2023/winrm-go"
  "fmt"
  "os"
)

endpoint := winrm.NewEndpoint("localhost", 5985, false, false, nil, nil, nil, 0)

params := DefaultParameters
params.TransportDecorator = func() Transporter { return &ClientNTLM{} }

client, err := NewClientWithParameters(endpoint, "test", "test", params)
if err != nil {
	panic(err)
}

_, err := client.RunWithInput("ipconfig", os.Stdout, os.Stderr, os.Stdin)
if err != nil {
	panic(err)
}

```

Passing a TransportDecorator also permit to use Kerberos authentication

```go
package main
import (
  "os"
  "fmt"
  "github.com/marslin2023/winrm-go"
)

endpoint := winrm.NewEndpoint("srv-win", 5985, false, false, nil, nil, nil, 0)

params := winrm.DefaultParameters
params.TransportDecorator = func() Transporter {
        return &winrm.ClientKerberos{
		Username: "test",
		Password: "s3cr3t",
		Hostname: "srv-win",
		Realm: "DOMAIN.LAN",
		Port: 5985,
		Proto: "http",
		KrbConf: "/etc/krb5.conf",
		SPN: fmt.Sprintf("HTTP/%s", hostname),
	}
}

client, err := NewClientWithParameters(endpoint, "test", "s3cr3t", params)
if err != nil {
        panic(err)
}

ctx, cancel := context.WithCancel(context.Background())
defer cancel()
_, err := client.RunWithContextWithInput(ctx, "ipconfig", os.Stdout, os.Stderr, os.Stdin)
if err != nil {
        panic(err)
}

```


By passing a Dial in the Parameters struct it is possible to use different dialer (e.g. tunnel through SSH)

```go
package main
     
 import (
    "github.com/marslin2023/winrm-go"
    "golang.org/x/crypto/ssh"
    "os"
 )
 
 func main() {
 
    sshClient, err := ssh.Dial("tcp","localhost:22", &ssh.ClientConfig{
        User:"ubuntu",
        Auth: []ssh.AuthMethod{ssh.Password("ubuntu")},
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
    })
 
    endpoint := winrm.NewEndpoint("other-host", 5985, false, false, nil, nil, nil, 0)
 
    params := winrm.DefaultParameters
    params.Dial = sshClient.Dial
 
    client, err := winrm.NewClientWithParameters(endpoint, "test", "test", params)
    if err != nil {
        panic(err)
    }
 
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    _, err = client.RunWithContextWithInput(ctx, "ipconfig", os.Stdout, os.Stderr, os.Stdin)
    if err != nil {
        panic(err)
    }
 }

```


For a more complex example, it is possible to call the various functions directly:

```go
package main

import (
  "github.com/marslin2023/winrm-go"
  "fmt"
  "bytes"
  "os"
)

stdin := bytes.NewBufferString("ipconfig /all")
endpoint := winrm.NewEndpoint("localhost", 5985, false, false,nil, nil, nil, 0)
client , err := winrm.NewClient(endpoint, "Administrator", "secret")
if err != nil {
	panic(err)
}
shell, err := client.CreateShell()
if err != nil {
  panic(err)
}
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
var cmd *winrm.Command
cmd, err = shell.ExecuteWithContext(ctx, "cmd.exe")
if err != nil {
  panic(err)
}

go io.Copy(cmd.Stdin, stdin)
go io.Copy(os.Stdout, cmd.Stdout)
go io.Copy(os.Stderr, cmd.Stderr)

cmd.Wait()
shell.Close()
```

For using HTTPS authentication with x 509 cert without checking the CA
```go
package main

import (
    "github.com/marslin2023/winrm-go"
    "log"
    "os"
)

func main() {
    clientCert, err := os.ReadFile("/home/example/winrm_client_cert.pem")
    if err != nil {
        log.Fatalf("failed to read client certificate: %q", err)
    }

    clientKey, err := os.ReadFile("/home/example/winrm_client_key.pem")
    if err != nil {
        log.Fatalf("failed to read client key: %q", err)
    }

    winrm.DefaultParameters.TransportDecorator = func() winrm.Transporter {
        // winrm https module
        return &winrm.ClientAuthRequest{}
    }

    endpoint := winrm.NewEndpoint(
        "192.168.100.2", // host to connect to
        5986,            // winrm port
        true,            // use TLS
        true,            // Allow insecure connection
        nil,             // CA certificate
        clientCert,      // Client Certificate
        clientKey,       // Client Key
        0,               // Timeout
    )
    client, err := winrm.NewClient(endpoint, "Administrator", "")
    if err != nil {
        log.Fatalf("failed to create client: %q", err)
    }
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    _, err = client.RunWithContext(ctx, "whoami", os.Stdout, os.Stderr)
    if err != nil {
        log.Fatalf("failed to run command: %q", err)
    }
}
```

Note: canceling the `context.Context` passed as first argument to the various
functions of the API will not cancel the HTTP requests themselves, it will
rather cause a running command to be aborted on the remote machine via a call to
`command.Stop()`.

## Developing on WinRM

If you wish to work on `winrm` itself, you'll first need [Go](http://golang.org)
installed (version 1.5+ is _required_). Make sure you have Go properly installed,
including setting up your [GOPATH](http://golang.org/doc/code.html#GOPATH).

For some additional dependencies, Go needs [Mercurial](http://mercurial.selenic.com/)
and [Bazaar](http://bazaar.canonical.com/en/) to be installed.
Winrm itself doesn't require these, but a dependency of a dependency does.

Next, clone this repository into `$GOPATH/src/github.com/marslin2023/winrm-go` and
then just type `make`.

You can run tests by typing `make test`.

If you make any changes to the code, run `make format` in order to automatically
format the code according to Go standards.

When new dependencies are added to winrm you can use `make updatedeps` to
get the latest and subsequently use `make` to compile.
