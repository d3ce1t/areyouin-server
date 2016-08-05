package utils

import "container/list"

type LinkedHashMap struct {
	list.List
	mapKeyToItem map[string]*list.Element
	mapItemToKey map[*list.Element]string
}

func NewLinkedHashMap() *LinkedHashMap {
	return &LinkedHashMap{
		mapKeyToItem: make(map[string]*list.Element),
		mapItemToKey: make(map[*list.Element]string),
	}
}

func (l *LinkedHashMap) KeyForItem(item *list.Element) string {
	if key, exist := l.mapItemToKey[item]; exist {
		return key
	}
	return ""
}

func (l *LinkedHashMap) PushFrontWithCollapseKey(key string, value interface{}) *list.Element {

	listItem, exist := l.mapKeyToItem[key]

	if exist {
		listItem.Value = value
		l.MoveToFront(listItem)
	} else {
		listItem = l.PushFront(value)
		l.mapKeyToItem[key] = listItem
		l.mapItemToKey[listItem] = key
	}

	return listItem
}

func (l *LinkedHashMap) PushBackWithCollapseKey(key string, value interface{}) *list.Element {

	listItem, exist := l.mapKeyToItem[key]

	if exist {
		listItem.Value = value
		l.MoveToBack(listItem)
	} else {
		listItem = l.PushBack(value)
		l.mapKeyToItem[key] = listItem
		l.mapItemToKey[listItem] = key
	}

	return listItem
}

func (l *LinkedHashMap) RemoveItemWithCollapseKey(key string) interface{} {

	if listItem, exist := l.mapKeyToItem[key]; exist {
		l.Remove(listItem)
		delete(l.mapKeyToItem, key)
		delete(l.mapItemToKey, listItem)
		return listItem.Value
	}

	return nil
}

func (l *LinkedHashMap) RemoveCollapseKey(key string) bool {
	if listItem, exist := l.mapKeyToItem[key]; exist {
		delete(l.mapKeyToItem, key)
		delete(l.mapItemToKey, listItem)
		return true
	}
	return false
}

func (l *LinkedHashMap) Remove(e *list.Element) interface{} {

	key, exist := l.mapItemToKey[e]

	if exist {
		delete(l.mapKeyToItem, key)
		delete(l.mapItemToKey, e)
	}

	return l.List.Remove(e)
}
