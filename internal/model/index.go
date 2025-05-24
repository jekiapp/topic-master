package model

import "fmt"

type Index struct {
	Name    string
	Pattern string
	Type    IndexType
}

type IndexType func(a, b string) bool

type IndexItem struct {
	Name string
	Type IndexType
}

func NewIndexes(tableName string, indexItems ...IndexItem) []Index {
	indexes := make([]Index, len(indexItems))
	for i, indexItem := range indexItems {
		indexes[i] = Index{
			Name:    tableName + ":" + indexItem.Name,
			Type:    indexItem.Type,
			Pattern: fmt.Sprintf("%s:*:%s", tableName, indexItem.Name),
		}
	}
	return indexes
}
