package ecs

func InstallShadowsocks() string {
	return "apt-get -y install wget && wget https://bootstrap.pypa.io/get-pip.py && python get-pip.py && pip install shadowsocks && echo '{ \"server\": \"0.0.0.0\", \"server_port\": 80, \"password\": \"123456\", \"timeout\": 300, \"method\": \"aes-256-cfb\" }' > /etc/shadowsocks.json && ssserver -c /etc/shadowsocks.json -d start"
}

func InstallUnixDev() string {
	return `
	curl -sL https://raw.githubusercontent.com/iamjinlei/env/master/unix_dev.sh | bash;
	echo -e "\n\n\n\n\n"
	echo -e "*****************************\n"
	echo -e "*   Install Unix Dev done   *\n"
	echo -e "*****************************\n"
	`
}

func InstallEthDev() string {
	return "curl -sL https://raw.githubusercontent.com/iamjinlei/env/master/unix_eth.sh | bash"
}

func RunProxy() string {
	return "mkdir -p $GOPATH/src/github.com/iamjinlei && cd $GOPATH/src/github.com/iamjinlei && git clone https://github.com/iamjinlei/ecs.git && go run ecs/proxy/main.go"
}
