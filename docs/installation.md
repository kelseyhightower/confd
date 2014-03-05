# Installation

### Binary Download

Currently confd ships binaries for OS X and Linux 64bit systems. You can download the latest release from [GitHub](https://github.com/kelseyhightower/confd/releases)

> Note: You don't need Go installed to use confd unless you plan to build from source.

Download confd version 0.2.0 using wget from the command line:
* OSX
```Bash
wget https://github.com/kelseyhightower/confd/releases/download/v0.2.0/confd_0.2.0_darwin_amd64.zip
```
* LINUX
```Bash
wget -O confd_0.2.0_linux_amd64.tar.gz https://github.com/kelseyhightower/confd/releases/download/v0.2.0/confd_0.2.0_linux_amd64.tar.gz
```

Unzip the confd package.
* OSX
```Bash
unzip confd_0.2.0_darwin_amd64.zip
```
* LINUX
```Bash
tar -zxvf confd_0.2.0_linux_amd64.tar.gz
```

Copy the confd binary to a bin directory in your path.

```Bash
sudo mv confd /usr/local/bin/confd
```

### Next Steps
Get up and running with the [Quick Start Guide](https://github.com/kelseyhightower/confd#quick-start).
