# huskie

export PORT=6000

Server

huskie harness --bind ${PORT}


Client

export HUSKIE_URL="https://huskie.run.aws-usw02-pr.ice.predix.io/tunnel"
export HUSKIE_URL="http://localhost:${PORT}/tunnel"

huskie pub --url $HUSKIE_URL
huskie mush --url $HUSKIE_URL


Create tunnels via two (pub) clients from p1 to p0

ssh:
/link p1:p0 {"remote":"<host>:22", "local":":60022"}

ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -p 60022 <user>@localhost
scp -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -P 60022 <user>@localhost

git:
/link p1:p0 {"remote":"github.com:22", "local":":61022"}

env GIT_SSH_COMMAND="ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no" git clone ssh://git@localhost:61022/<org>/<repo>.git

database:
/link p1:p0 {"remote":"<host>:1561", "local":":61561"}

