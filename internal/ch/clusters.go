package ch

import "context"

// ClusterNodeCount 为 system.clusters 节点列表需要的列数。
const clusterCols = 6

// Clusters 读取 system.clusters 下的集群节点列表。
func (c *Client) Clusters(ctx context.Context) ([]Node, error) {
	res, err := c.Query(ctx,
		"SELECT cluster, shard_num, replica_num, host_name, host_address, port FROM system.clusters ORDER BY cluster, shard_num, replica_num",
		0)
	if err != nil {
		return nil, err
	}
	nodes := make([]Node, 0, len(res.Rows))
	for _, row := range res.Rows {
		if len(row) < clusterCols {
			continue
		}
		nodes = append(nodes, Node{
			Cluster:  FormatValue(row[0]),
			Shard:    FormatValue(row[1]),
			Replica:  FormatValue(row[2]),
			HostName: FormatValue(row[3]),
			Address:  FormatValue(row[4]),
			Port:     atoi(FormatValue(row[5])),
		})
	}
	return nodes, nil
}
