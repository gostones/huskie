# huskie

export PORT=6000

Server

huskie harness --bind ${PORT}


Client

export HUSKIE_URL="https://huskie.run.aws-usw02-pr.ice.predix.io/tunnel"
export HUSKIE_URL="http://localhost:${PORT}/tunnel"

huskie pub --url $HUSKIE_URL
huskie mush --url $HUSKIE_URL
