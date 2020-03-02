set -eumo pipefail

readonly node_count=5
readonly host_name=localhost

tmpdir=$(mktemp -d)
cd "$tmpdir"

trap cleanup EXIT QUIT
cleanup() {
	local err=$?
	local pid

	for pid in $nodes
	do
		kill -9 $pid
	done

	for pid in $nodes
	do
		set +e
		wait $pid 2>/dev/null
		local ret=$?
		set -e

		if [ $ret -ne 137 ]
		then
			err=$ret
		fi
	done

	rm -rf "$tmpdir"

	exit $err
}

fail() {
	echo $@
	exit 1
}

readonly port_base=$((RANDOM + 1024))
readonly port_top=$((port_base + 2*node_count - 1))
nodes=''
publics=''
start_nodes() {
	local loader=random
	if [ $# -eq 1 ]
	then
		loader="file-loader $1"
	fi

	[ -n "$nodes" ] && ( echo nodes already started; exit 1 )

	local port

	for port in $(seq $port_base 2 $port_top)
	do
		local node_conf=$(server new $host_name:{$port,$((port+1))})
		publics+=" $(echo "$node_conf" | awk -F \" '/^\s+Public\s*=/ {print $2}')"

		echo "$node_conf" |
				server data-provider new $loader |
					server data-provider set-neutralizer minimum-results-size 0 |
				server computing-node new |
				server verifying-node new |
				DEBUG_COLOR=true server run &
		nodes+=" $!"
	done

	for port in $(seq $port_base $port_top)
	do
		while ! nc -q 0 localhost $port < /dev/null
		do
			sleep 0.1
		done
	done
}

get_nodes() {
	[ -z "$nodes" ] && ( echo asking roster of stopped nodes; exit 1 )

	local port=$port_base
	for public in $publics
	do
		echo $host_name:$port $public
		: $((port += 2))
	done
}

get_client() {
	[ -z "$nodes" ] && ( echo asking for client to stopped nodes; exit 1 )

	echo $host_name:$((port_base+1))
}

client_gen_network() {
	local pipe=$(
		echo -n client network new
		for n in $(get_nodes | tr ' ' ,)
		do
			echo -n " | client network add-node $(echo $n | tr , ' ')"
		done
		echo -n " | client network set-client $(get_client)"
	)
	eval "$pipe"
}
