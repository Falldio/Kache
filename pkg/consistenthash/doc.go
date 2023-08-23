// package consistenthash allows nodes in the cluster to fetch cache
// from other nodes. Nodes are arranged in a hash ring and kvs are allocated
// to them accordingly. If the node we are communicating doesn't have the kv we
// are querying, it automatically fecthes the data from other nodes.
package consistenthash
