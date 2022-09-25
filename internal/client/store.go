package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

const (
	storeDir         = ".homecloud"
	clusterFileName  = "cluster.json"
	nodeFileName     = "node.json"
	osConfigFileName = "hcos.yaml"
)

type ErrNotFound struct {
	s string
}

func (e *ErrNotFound) Error() string {
	return e.s
}

type Store struct {
	rootDir string
}

func LoadOrCreate(rootDir string) (*Store, error) {
	if rootDir == "" {
		rootDir = getDefaultDir()
	}
	s := &Store{rootDir: rootDir}
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) GetCluster(name string) (Cluster, error) {
	path := s.clusterPath(name)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Cluster{}, &ErrNotFound{fmt.Sprintf("cluster %q not found", name)}
		}
		return Cluster{}, err
	}
	var cluster Cluster
	if err := json.Unmarshal(data, &cluster); err != nil {
		return Cluster{}, err
	}
	return cluster, nil
}

func (s *Store) ListClusters() ([]Cluster, error) {
	dir := filepath.Join(s.rootDir, "clusters")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	clusters := make([]Cluster, 0, len(entries))
	for _, entry := range entries {
		cluster, err := s.GetCluster(entry.Name())
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)
	}
	return clusters, nil
}

func (s *Store) SaveCluster(cluster Cluster) error {
	path := s.clusterPath(cluster.Name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(cluster)
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	return nil
}

func (s *Store) clusterPath(name string) string {
	return filepath.Join(s.rootDir, "clusters", name, clusterFileName)
}

func (s *Store) GetNode(clusterName, name string) (Node, error) {
	dir := s.nodeDir(clusterName, name)
	path := filepath.Join(dir, nodeFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Node{}, &ErrNotFound{fmt.Sprintf("node %q not found in cluster %q", name, clusterName)}
		}
		return Node{}, err
	}
	var node Node
	if err := json.Unmarshal(data, &node); err != nil {
		return Node{}, err
	}
	osCfgPath := filepath.Join(dir, osConfigFileName)
	osCfgData, err := os.ReadFile(osCfgPath)
	if err != nil {
		return Node{}, err
	}
	if err := yaml.Unmarshal(osCfgData, &node.OSConfig); err != nil {
		return Node{}, err
	}
	return node, nil
}

func (s *Store) ListNodes(clusterName string) ([]Node, error) {
	dir := filepath.Join(filepath.Dir(s.clusterPath(clusterName)), "nodes")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	nodes := make([]Node, 0, len(entries))
	for _, entry := range entries {
		node, err := s.GetNode(clusterName, entry.Name())
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func (s *Store) SaveNode(clusterName string, node Node) error {
	dir := s.nodeDir(clusterName, node.Name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	nodeData, err := json.Marshal(node)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, nodeFileName), nodeData, 0644); err != nil {
		return err
	}

	var osCfgData bytes.Buffer
	enc := yaml.NewEncoder(&osCfgData)
	enc.SetIndent(2)
	if err := enc.Encode(node.OSConfig); err != nil {
		return err
	}
	if err := enc.Close(); err != nil {
		return err
	}
	// OS config contains sensitive data, so we need to make sure it's not readable by other users.
	return os.WriteFile(filepath.Join(dir, osConfigFileName), osCfgData.Bytes(), 0600)
}

func (s *Store) nodeDir(clusterName, name string) string {
	return filepath.Join(filepath.Dir(s.clusterPath(clusterName)), "nodes", name)
}

func getDefaultDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "" // Current working directory.
	}
	return filepath.Join(homeDir, storeDir)
}
