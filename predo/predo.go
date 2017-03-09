package predo

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/mafengwo/confd/log"
)

// 找出需要用命名空间替换的一对对的模版文件(需要做验证 文件名的开头是以服务名开头的)
type tomlConfig struct {
	Ini   info
	Nginx info
}

type info struct {
	Tomlx []string
	Tmplx []string
}

/*
 * 完成confd前的任务：根据etcd中的namespace将对应的标准配置文件和标准模版文件中的变量替换掉
 * 1.根据特殊的命名规则检查标准配置文件和模版文件是否存在
 * 2.用namespace替换变量生成一对临时的配置文件和模版文件，判断是否一样。如果不一样，替换成新的配置文件和模版文件
 *
 * 命令规范:
 * namespace - es-XXX-data, es-XXX-master, redis, memcached. (注：es-XXX-data和es-XXX-master只需生成一对配置文件和模版文件)(服务-业务-data || 服务-业务-master)
 * 标准配置文件 - es.tomlx, redis.tomlx(命名空间的替换词为 __NS__)
 * 标准模版文件 - es.tmplx, redis.tmplx(命名空间的替换词为 __NS__)
 * 配置文件 - es-XXX.toml
 * 模版文件 - es-XXX-tmpl
 */

// MainProcess 主程序入口
func MainProcess(configDir, templateDir string, namespcae []string) {
	var newNamespace, namespcaeArr []string
	var config tomlConfig
	var tomlxFile, tmplxFile string
	// 处理namespace 去掉后缀同时去重
	for _, value := range namespcae {
		namespcaeArr = strings.Split(value, "-")
		if len(namespcaeArr) == 3 || len(namespcaeArr) == 2 {
			newNamespace = append(newNamespace, namespcaeArr[0]+"-"+namespcaeArr[1])
		} else {
			newNamespace = append(newNamespace, value)
		}
	}
	newNamespace = Rmduplicate(&newNamespace)

	_, err := toml.DecodeFile(filepath.Join(configDir, "batch.ini"), &config)
	if err != nil {
		log.Info("predo: parse batch.ini faild error= %s", err.Error())
	}
	//进行验证 要保证tompx和tmplx是成对出现的
	if len(config.Ini.Tmplx) != len(config.Ini.Tomlx) || len(config.Nginx.Tmplx) != len(config.Nginx.Tomlx) {
		log.Info("predo: batch.ini toplx and tmplx number is wrong! ")
		return
	}
	//处理Ini模块
	for i := 0; i < len(config.Ini.Tmplx); i++ {
		tomlxFile = config.Ini.Tomlx[i]
		tmplxFile = config.Ini.Tmplx[i]
		for _, item := range newNamespace {
			if check(configDir, templateDir, tomlxFile, tmplxFile, item) {
				//处理配置文件
				handleNamespace(configDir, ".toml", item, tomlxFile, "ini")
				//处理模版文件
				handleNamespace(templateDir, ".tmpl", item, tmplxFile, "ini")
			}
		}
	}
	//处理Nginx模块
	for i := 0; i < len(config.Nginx.Tmplx); i++ {
		tomlxFile = config.Nginx.Tomlx[i]
		tmplxFile = config.Nginx.Tmplx[i]
		for _, item := range newNamespace {
			if check(configDir, templateDir, tomlxFile, tmplxFile, item) {
				//处理配置文件
				handleNamespace(configDir, ".toml", item, tomlxFile, "nginx")
				//处理模版文件
				handleNamespace(templateDir, ".tmpl", item, tmplxFile, "nginx")
			}
		}
	}
}

// 检验标准的配置文件是否存在以及和namaspace的关系
func check(configDir, templateDir, tomlxFile, tmplxFile, namespace string) bool {
	//判断路径是否存在
	tomlxPath := filepath.Join(configDir, tomlxFile)
	tmplxPath := filepath.Join(templateDir, tmplxFile)
	if !pathExists(tomlxPath) || !pathExists(tmplxPath) {
		log.Info("predo: not Exists tomlxPath=%s, tmplxPath=%s, namespace=%s", tomlxPath, tmplxPath, namespace)
		return false
	}

	//判断文件名与namespace的关系,开头必须要以服务名开头
	serviceName := strings.Split(namespace, "-")[0]
	tomlxName := strings.Split(strings.TrimSuffix(tomlxFile, filepath.Ext(tomlxFile)), "-")[0]
	tmplxName := strings.Split(strings.TrimSuffix(tmplxFile, filepath.Ext(tmplxFile)), "-")[0]
	if serviceName != tomlxName || serviceName != tmplxName || tmplxName != tomlxName {
		return false
	}
	log.Info("predo: %s namespace has config(%s) and template(%s)", namespace, tomlxFile, tmplxFile)
	return true

}

// 根据namespace判断是否存在标准的配置文件和标准的模版文件 此处的namespace已经是转化为es-XXX类似的.标准配置文件(es.tomlx)和标准模版文件(es.tmplx)
func isExists(configDir, templateDir, namespcae string) bool {
	defaultConfigPath := filepath.Join(configDir, strings.Split(namespcae, "-")[0]+".tomlx")
	defaultTemplatePath := filepath.Join(templateDir, strings.Split(namespcae, "-")[0]+".tmplx")
	if pathExists(defaultConfigPath) && pathExists(defaultTemplatePath) {
		log.Info("predo: %s namespace has config and template", namespcae)
		return true
	}
	return false
}

// 用namespace替换变量生成一对临时的配置文件和模版文件，判断是否一样。如果不一样，替换成新的配置文件和模版文件
func handleNamespace(dir, configSuffix, namespcae, standardFile, ext string) {
	tempConfigFile, err := ioutil.TempFile(dir, "temp")
	if err != nil {
		log.Info("predo: create temp config(%s) faild ", namespcae)
		return
	}
	defer os.Remove(tempConfigFile.Name())

	defaultConfigPath := filepath.Join(dir, standardFile)
	configPath := filepath.Join(dir, namespcae+"-"+ext+configSuffix)
	err = replace2File(tempConfigFile.Name(), defaultConfigPath, namespcae)
	if err != nil {
		log.Info("predo: replace2File error(%s)", err.Error())
		return
	}
	if !pathExists(configPath) {
		err := os.Rename(tempConfigFile.Name(), configPath)
		if err != nil {
			log.Info("predo: (%s)namespace rename1 faild ", namespcae)
		}
		//存在 作对比，文件内容没变化，不做替换
	} else {
		md5sum1, _ := run("md5sum " + configPath + " | awk '{print $1}'")
		md5sum2, _ := run("md5sum " + tempConfigFile.Name() + " | awk '{print $1}'")
		if md5sum1 != md5sum2 {
			err := os.Rename(tempConfigFile.Name(), configPath)
			if err != nil {
				log.Info("predo: (%s)namespace rename2 faild ", namespcae+configSuffix)
			}
		} else {
			log.Info("predo: (%s)namespace file is same", namespcae+configSuffix)
		}
	}
}

// 替换变量 生成配置文件和模版文件
func replace2File(tempFile, file, namespace string) error {
	cmd := `sed "s/\${NS}/` + namespace + `/g" ` + file + ` > ` + tempFile
	_, err := run(cmd)
	return err
}

// Run shell脚本
func run(shell string) (string, error) {
	cmd := exec.Command("/bin/sh", "-c", shell)
	out, err := cmd.Output()
	return string(out), err
}

// pathExists 判断路径是否存在(绝对路径)
func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

// Rmduplicate unique slice
func Rmduplicate(list *[]string) []string {
	var x = []string{}
	for _, i := range *list {
		if len(x) == 0 {
			x = append(x, i)
		} else {
			for k, v := range x {
				if i == v {
					break
				}
				if k == len(x)-1 {
					x = append(x, i)
				}
			}
		}
	}
	return x
}
