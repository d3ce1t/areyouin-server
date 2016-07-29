#/usr/bin/bash
idgen_pkg="peeple/areyouin/idgen"
utils_pkg="peeple/areyouin/utils"
api_pkg="peeple/areyouin/api"
cqldao_pkg="peeple/areyouin/cqldao"
model_pkg="peeple/areyouin/model"
protocol_core_pkg="peeple/areyouin/protocol/core"
protocol_pkg="peeple/areyouin/protocol"
facebook_pkg="peeple/areyouin/facebook"
webhook_pkg="peeple/areyouin/webhook"
images_server_pkg="peeple/areyouin/images_server"
server_shell_pkg="peeple/areyouin/server/shell"
server="peeple/areyouin/server"

function build_and_install {
  echo "Build $1"
  go build $1
  echo "Install $1"
  go install $1
}

build_and_install $idgen_pkg
build_and_install $utils_pkg
build_and_install $api_pkg
build_and_install $cqldao_pkg
build_and_install $model_pkg
build_and_install $protocol_core_pkg
build_and_install $protocol_pkg
build_and_install $facebook_pkg
build_and_install $webhook_pkg
build_and_install $images_server_pkg
build_and_install $server_shell_pkg
cd server
build_and_install $server
cd ..
