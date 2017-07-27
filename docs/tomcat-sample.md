# confd for Apache Tomcat
If you administrate an [Apache Tomcat](http://tomcat.apache.org/) you usually need to edit multiple config files and set some environment variables. 
[confd](https://github.com/kelseyhightower/confd) can help here especially if you have multiple Tomcats in a cluster. Configuring Tomcat is an interesting sample, because it needs multiple config files and some environment variables.

Important configuration files of Tomcat are 
- server.xml: e.g. configure here jvmRoute for load balancing, ports etc.
- tomcat-users.xml: configure access rights for manager webapp

A frequently used environment variable is CATALINA_OPTS: Here you can define memory settings

## server.xml
In this case we simply want to use the hostname for jvmRoute. Because it is not possible to execute a Unix command within a toml template, we use an environment variable "HOSTNAME" instead.

Copy the file conf/server.xml from your Tomcat installation to /etc/confd/templates/server.xml.tmpl and edit "Engine tag" and add jvmRoute
```
<Engine name="Catalina" defaultHost="localhost" jvmRoute="{{getenv "HOSTNAME"}}">
```

Create /etc/confd/conf.d/server.xml.toml
```
[template]
src = "server.xml.tmpl"
dest = "/usr/local/tomcat/conf/server.xml"

check_cmd = "/usr/local/tomcat/bin/catalina.sh configtest"
reload_cmd = "/usr/local/tomcat/bin/catalina.sh stop -force && /usr/local/tomcat/bin/catalina.sh start"
```

## tomcat-users.xml
We want to get the Tomcat-manager's credentials from a central configuration repository. 
Copy the file conf/tomcat-users.xml from your Tomcat installation to /etc/confd/templates/tomcat-users.xml.tmpl
Add the following lines:
```
<role rolename="tomcat"/>
<role rolename="manager-gui"/>
<user username="{{getv "/user"}}" password="{{getv "/password"}}" roles="tomcat,manager-gui"/>
```

Create a /etc/confd/conf.d/tomcat-users.xml.toml
```
[template]
prefix = "tomcat"
keys = [
  "user",
  "password"
]

src = "tomcat-users.xml.tmpl"
dest = "/usr/local/tomcat/conf/tomcat-users.xml"

reload_cmd = "/usr/local/tomcat/bin/catalina.sh stop -force && /usr/local/tomcat/bin/catalina.sh start"
```

 
## catalina.sh
File catalina.sh is the startscript for Tomcat. If confd should set memory settings like Xmx or Xms, we could either create a catalina.sh.tmpl and proceed like above or we can try to use environment variables and leave catalina.sh untouched. Leaving catalina.sh untouched is preferred here. Because it is not possible to use environment variables within toml files, we need to write a minimal shell script that passes CATALINA_OPTS variable.

Create the file /etc/confd/conf.d/catalina_start.sh.tmpl
```
#!/bin/sh
CATALINA_OPTS="-Xms{{getv "/Xms"}} -Xmx{{getv "/Xmx"}}" /usr/local/tomcat/bin/catalina.sh start
```

Create the file /etc/confd/conf.d/catalina_start.sh.toml
```
[template]
prefix = "tomcat"
keys = [
  "Xmx",
  "Xms"
]

src = "catalina_start.sh.tmpl"
dest = "/usr/local/tomcat/bin/catalina_start.sh"
mode = "0775"

reload_cmd = "/usr/local/tomcat/bin/catalina.sh stop -force && /usr/local/tomcat/bin/catalina_start.sh start"
```

Finally we need to replace in all above templates ```catalina.sh start``` by ```catalina_start.sh start```

## test it
Follow conf documentation and test it calling ```confd -onetime``` or you try the complete sample in a Docker and/or Vagrant environment [here](https://github.com/muenchhausen/tomcat-confd)

