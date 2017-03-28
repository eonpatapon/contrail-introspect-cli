// An introspect page is mapped into a Collection which is basically a
// list of Element.

package collection

import "fmt"

import "github.com/jbowtie/gokogiri/xml"

// A Collection describes and contains a list of Elements.
type Collection struct {
	Url     string
	descCol DescCollection
	doc     *xml.XmlDocument
	// The node containing the whole XML, for instance the whole
	// loaded XML Page.
	rootNode xml.Node
	elements []Element
}

type Elements []Element

type Element struct {
	Node xml.Node
	desc DescElement
}

// This contains informations to generate and query a Collection
type DescCollection struct {
	// Names of arguments required from the user to get datas from
	// introspect
	PageArgs []string
	// A function that takes the list of arguments specified by
	// the user
	PageBuilder (func([]string) Sourcer)
	// The root Xpath
	BaseXpath string
	// Description of Collection's Elements
	DescElt DescElement
	// Name of the attribute used to search in the collection
	PrimaryField string
}

type DescElement struct {
	// Xpath used to generate the short version of an element
	ShortDetailXpath string
	// Used to generate the long version of an element
	LongDetail LongFormatter
}

func (col *Collection) Init() {
	ss, _ := col.rootNode.Search(col.descCol.BaseXpath + "/*")
	col.elements = make([]Element, len(ss))
	for i, s := range ss {
		col.elements[i] = Element{Node: s, desc: col.descCol.DescElt}
	}
}

func (col *Collection) Close() {
	// Collection must be closed manually to free memory
	// https://github.com/moovweb/gokogiri/issues/36
	col.doc.Free()
}

// If key == "", the PrimaryField is used
func (col *Collection) SearchXpathFuzzy(key string, pattern string) string {
	if key == "" {
		key = col.descCol.PrimaryField
	}
	return col.descCol.BaseXpath + "/*/" + key + "[contains(text(),'" + pattern + "')]/.."
}

// If key == "", the PrimaryField is used
func (col *Collection) SearchXpathStrict(key string, pattern string) string {
	if key == "" {
		key = col.descCol.PrimaryField
	}
	return col.descCol.BaseXpath + "/*/" + key + "[text()='" + pattern + "']/.."
}

func (col *Collection) SearchFuzzy(pattern string) Elements {
	return col.Search(col.SearchXpathFuzzy, "", pattern)
}

func (col *Collection) SearchStrict(pattern string) Elements {
	return col.Search(col.SearchXpathStrict, "", pattern)
}

func (col *Collection) SearchStrictUniqueByKey(key string, pattern string) (Element, error) {
	res := col.Search(col.SearchXpathStrict, key, pattern)
	if len(res) != 1 {
		return Element{}, fmt.Errorf("Pattern %s should match exactly one element (instead of '%s')", pattern, res)
	}
	return res[0], nil
}

func (col *Collection) SearchFuzzyUnique(pattern string) (Element, error) {
	res := col.SearchFuzzy(pattern)
	if len(res) != 1 {
		return Element{}, fmt.Errorf("Pattern %s should match exactly one element (instead of '%s')", pattern, res)
	}
	return res[0], nil
}

func (col *Collection) SearchStrictUnique(pattern string) (Element, error) {
	res := col.SearchStrict(pattern)
	if len(res) != 1 {
		return Element{}, fmt.Errorf("Pattern %s should match exactly one element (instead of '%s')", pattern, res)
	}
	return res[0], nil
}

func (col *Collection) Search(searchPredicate func(string, string) string, key string, pattern string) Elements {
	ss, _ := col.rootNode.Search(searchPredicate(key, pattern))
	var elements []Element = make([]Element, len(ss))
	for i, s := range ss {
		elements[i] = Element{Node: s, desc: col.descCol.DescElt}
	}
	return Elements(elements)
}

func (e Element) GetField(field string) (string, error) {
	s, _ := e.Node.Search(fmt.Sprintf("%s/text()", field))
	for i, _ := range s {
		if s[i].String() != s[0].String() {
			return "", fmt.Errorf("All fields values must be equal (values: %s)", s)
		}
	}
	if len(s) < 1 {
		return "", fmt.Errorf("Field %s has not be found.", field)
	}
	return s[0].String(), nil
}
