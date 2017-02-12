package reflector

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tkrajina/go-reflector/reflector/tmp"
)

type Address struct {
	Street string `tag:"be" tag2:"1,2,3"`
	Number int    `tag:"bi"`
}

type Person struct {
	Name string `tag:"bu"`
	Address
}

func (p Person) Add(a, b, c int) int     { return a + b + c }
func (p *Person) Substract(a, b int) int { return a - b }
func (p Person) ReturnsError(err bool) (string, *int, error) {
	i := 2
	if err {
		return "", nil, errors.New("Error here!")
	}
	return "jen", &i, nil
}

type CustomType int

func (ct CustomType) Method1() string { return "yep" }
func (ct *CustomType) Method2() int   { return 7 }

func (p Person) Hi(name string) string {
	return fmt.Sprintf("Hi %s my name is %s", name, p.Name)
}

type Company struct {
	Address
	Number int `tag:"bi"`
}

// TestDoubleFields checks that there are no "double" fields (or fields shadowing)
func TestDoubleFields(t *testing.T) {
	structs := []interface{}{
		Obj{},
		ObjField{},
		ObjMethod{},
		ObjMetadata{},
		ObjFieldMetadata{},
		ObjMethodMetadata{},
	}
	for _, s := range structs {
		double := New(s).FindDoubleFields()
		assert.Equal(t, 0, len(double))
	}
}

func TestString(t *testing.T) {
	assert.Equal(t, "reflector.Person", New(Person{}).String())
	assert.Equal(t, "*reflector.Person", New(&Person{}).String())
	assert.Equal(t, "nil", New(nil).String())
	assert.Equal(t, "int", New(1).String())
	var i int
	assert.Equal(t, "*int", New(&i).String())

	// Check that when we twice get a field, the metadata field is cached only once:
	assert.Equal(t, New(Person{}).Field("bu").ObjFieldMetadata, New(Person{}).Field("bu").ObjFieldMetadata)
	assert.Equal(t, New(&Person{}).Field("bu").ObjFieldMetadata, New(&Person{}).Field("bu").ObjFieldMetadata)
	assert.Equal(t, New(Person{}).Field("Address").ObjFieldMetadata, New(Person{}).Field("Address").ObjFieldMetadata)
	assert.Equal(t, New(&Person{}).Field("bu").ObjFieldMetadata, New(&Person{}).Field("bu").ObjFieldMetadata)
}

func TestNilStringPtr(t *testing.T) {
	assert.Equal(t, "*string", New((*string)(nil)).String())
	assert.Equal(t, 0, len(New((*string)(nil)).Fields()))
	assert.Equal(t, 0, len(New((*string)(nil)).Methods()))

	v, err := New((*string)(nil)).Field("Bu").Get()
	assert.Nil(t, v)
	assert.NotNil(t, err)
}

func TestNilStructPtr(t *testing.T) {
	assert.Equal(t, "*reflector.Person", New((*Person)(nil)).String())
	assert.Equal(t, 2, len(New((*Person)(nil)).Fields()))
	assert.Equal(t, 4, len(New((*Person)(nil)).Methods()))

	err := New((*Person)(nil)).Field("Number").Set(17)
	assert.NotNil(t, err)

	v, err := New((*Person)(nil)).Field("Bu").Get()
	assert.Nil(t, v)
	assert.NotNil(t, err)
}

func TestListFieldsFlattened(t *testing.T) {
	p := Person{}
	obj := New(p)

	assert.True(t, obj.IsValid())
	assert.False(t, obj.IsPtr())
	assert.True(t, obj.IsStructOrPtrToStruct())

	fields := obj.FieldsFlattened()
	assert.Equal(t, 3, len(fields))
	assert.Equal(t, fields[0].Name(), "Name")
	assert.Equal(t, fields[1].Name(), "Street")
	assert.Equal(t, fields[2].Name(), "Number")

	assert.True(t, obj.Field("Name").IsValid())

	kind := obj.Field("Name").Kind()
	assert.Equal(t, reflect.String, kind)

	kind = obj.Field("BuName").Kind()
	assert.Equal(t, reflect.Invalid, kind)

	ty := obj.Field("Number").Type()
	assert.Equal(t, reflect.TypeOf(1), ty)

	ty = obj.Field("Istra").Type()
	assert.Nil(t, ty)
}

func TestListFields(t *testing.T) {
	p := Person{}
	obj := New(p)

	fields := obj.Fields()
	assert.Equal(t, len(fields), 2)
	assert.Equal(t, fields[0].Name(), "Name")
	assert.Equal(t, fields[1].Name(), "Address")
}

func TestListFieldsAll(t *testing.T) {
	p := Person{}
	obj := New(p)

	fields := obj.FieldsAll()
	assert.Equal(t, len(fields), 4)
	assert.Equal(t, fields[0].Name(), "Name")
	assert.Equal(t, fields[1].Name(), "Address")
	assert.Equal(t, fields[2].Name(), "Street")
	assert.Equal(t, fields[3].Name(), "Number")
}

func TestListFieldsAnonymous(t *testing.T) {
	p := Person{}
	obj := New(p)

	fields := obj.FieldsAnonymous()
	assert.Equal(t, 1, len(fields))
	assert.Equal(t, fields[0].Name(), "Address")
}

func TestListFieldsAllWithDoubleFields(t *testing.T) {
	obj := New(Company{})

	fields := obj.FieldsAll()
	assert.Equal(t, len(fields), 4)
	assert.Equal(t, fields[0].Name(), "Address")
	assert.Equal(t, fields[1].Name(), "Street")
	// Number is declared both in Company and in Address, so listed twice here:
	assert.Equal(t, fields[2].Name(), "Number")
	assert.Equal(t, fields[3].Name(), "Number")
}

func TestFindDoubleFields(t *testing.T) {
	obj := New(Company{})

	fields := obj.FindDoubleFields()
	assert.Equal(t, 1, len(fields))
	assert.Equal(t, fields[0], "Number")
}

func TestListFieldsOnPointer(t *testing.T) {
	p := &Person{}
	obj := New(p)

	assert.True(t, obj.IsPtr())
	assert.True(t, obj.IsStructOrPtrToStruct())

	fields := obj.Fields()
	assert.Equal(t, len(fields), 2)
	assert.Equal(t, fields[0].Name(), "Name")
	assert.Equal(t, fields[1].Name(), "Address")

	kind := obj.Field("Name").Kind()
	assert.Equal(t, reflect.String, kind)

	kind = obj.Field("BuName").Kind()
	assert.Equal(t, reflect.Invalid, kind)

	ty := obj.Field("Number").Type()
	assert.Equal(t, reflect.TypeOf(1), ty)

	ty = obj.Field("Istra").Type()
	assert.Nil(t, ty)
}

func TestListFieldsFlattenedOnPointer(t *testing.T) {
	p := &Person{}
	obj := New(p)

	fields := obj.FieldsFlattened()
	assert.Equal(t, len(fields), 3)
	assert.Equal(t, fields[0].Name(), "Name")
	assert.Equal(t, fields[1].Name(), "Street")
	assert.Equal(t, fields[2].Name(), "Number")
}

func TestNoFieldsNoCustomType(t *testing.T) {
	assert.Equal(t, len(New(CustomType(1)).Fields()), 0)
	ct := CustomType(2)
	assert.Equal(t, len(New(&ct).Fields()), 0)
}

func TestIsStructForCustomTypes(t *testing.T) {
	ct := CustomType(2)
	assert.False(t, New(CustomType(1)).IsPtr())
	assert.True(t, New(&ct).IsPtr())
	assert.False(t, New(CustomType(1)).IsStructOrPtrToStruct())
	assert.False(t, New(&ct).IsStructOrPtrToStruct())
}

func TestFieldValidity(t *testing.T) {
	assert.False(t, New(CustomType(1)).Field("jkljkl").IsValid())
	assert.False(t, New(Person{}).Field("street").IsValid())
	assert.True(t, New(Person{}).Field("Street").IsValid())
	assert.True(t, New(Person{}).Field("Number").IsValid())
	assert.True(t, New(Person{}).Field("Name").IsValid())
}

func TestSetFieldNonPointer(t *testing.T) {
	p := Person{}
	obj := New(p)
	assert.False(t, obj.IsPtr())

	field := obj.Field("Street")
	assert.False(t, field.IsSettable())

	err := field.Set("ulica")
	assert.Error(t, err)

	value, err := field.Get()
	assert.Nil(t, err)
	assert.Equal(t, "", value)

	assert.Equal(t, "", p.Street)

	street, err := obj.Field("Street").Get()
	assert.Nil(t, err)

	// This actually don't work because p is a struct and reflector is working on it's own copy:
	assert.Equal(t, "", street)

}

func TestSetField(t *testing.T) {
	p := Person{}
	obj := New(&p)
	assert.True(t, obj.IsPtr())

	field := obj.Field("Street")
	assert.True(t, field.IsSettable())

	err := field.Set("ulica")
	assert.Nil(t, err)

	value, err := field.Get()
	assert.Nil(t, err)
	assert.Equal(t, "ulica", value)

	assert.Equal(t, "ulica", p.Street)
}

func TestCustomTypeMethods(t *testing.T) {
	assert.Equal(t, len(New(CustomType(1)).Methods()), 1)
	ct := CustomType(1)
	assert.Equal(t, len(New(&ct).Methods()), 2)
}

func TestMethods(t *testing.T) {
	assert.Equal(t, len(New(Person{}).Methods()), 3)
	assert.Equal(t, len(New(&Person{}).Methods()), 4)

	// Check that when we twice get a field, the metadata field is cached only once:
	assert.Equal(t, New(Person{}).Method("Add").ObjMethodMetadata, New(Person{}).Method("Add").ObjMethodMetadata)
	assert.Equal(t, New(&Person{}).Method("Add").ObjMethodMetadata, New(&Person{}).Method("Add").ObjMethodMetadata)
}

func TestCallMethod(t *testing.T) {
	obj := New(&Person{})
	method := obj.Method("Add")
	res, err := method.Call(2, 3, 6)
	assert.Nil(t, err)
	assert.False(t, res.IsError())
	assert.Equal(t, len(res.Result), 1)
	assert.Equal(t, res.Result[0], 11)

	assert.True(t, method.IsValid())
	assert.Equal(t, len(method.InTypes()), 3)
	assert.Equal(t, len(method.OutTypes()), 1)

	sub, err := obj.Method("Substract").Call(5, 6)
	assert.Nil(t, err)
	assert.Equal(t, sub.Result, []interface{}{-1})
}

func TestCallInvalidMethod(t *testing.T) {
	obj := New(&Person{})
	method := obj.Method("AddAdddd")
	res, err := method.Call([]interface{}{2, 3, 6})
	assert.NotNil(t, err)
	assert.Nil(t, res)

	assert.Equal(t, len(method.InTypes()), 0)
	assert.Equal(t, len(method.OutTypes()), 0)
}

func TestMethodsValidityOnPtr(t *testing.T) {
	ct := CustomType(1)
	obj := New(&ct)

	assert.True(t, obj.IsPtr())

	assert.True(t, obj.Method("Method1").IsValid())
	assert.True(t, obj.Method("Method2").IsValid())

	{
		res, err := obj.Method("Method1").Call()
		assert.Nil(t, err)
		assert.Equal(t, res.Result, []interface{}{"yep"})
	}
	{
		res, err := obj.Method("Method2").Call()
		assert.Nil(t, err)
		assert.Equal(t, res.Result, []interface{}{7})
	}
}

func TestMethodsValidityOnNonPtr(t *testing.T) {
	obj := New(CustomType(1))

	assert.False(t, obj.IsPtr())

	assert.True(t, obj.Method("Method1").IsValid())
	// False because it's not a pointer
	assert.False(t, obj.Method("Method2").IsValid())

	{
		res, err := obj.Method("Method1").Call()
		assert.Nil(t, err)
		assert.Equal(t, res.Result, []interface{}{"yep"})
	}
	{
		_, err := obj.Method("Method2").Call()
		assert.NotNil(t, err)
	}
}

func testCallMethod(t *testing.T, callValue bool, lenResult int) bool {
	obj := New(&Person{})
	res, err := obj.Method("ReturnsError").Call(callValue)
	assert.Nil(t, err)
	assert.Equal(t, len(res.Result), lenResult)
	return res.IsError()
}

func TestCallMethodWithoutErrResult(t *testing.T) {
	isErr := testCallMethod(t, true, 3)
	assert.True(t, isErr)
}

func TestCallMethodWithErrResult(t *testing.T) {
	isErr := testCallMethod(t, false, 3)
	assert.False(t, isErr)
}

func TestTag(t *testing.T) {
	obj := New(&Person{})
	tag, err := obj.Field("Street").Tag("invalid")
	assert.Nil(t, err)
	assert.Equal(t, len(tag), 0)
}

func TestInvalidTag(t *testing.T) {
	obj := New(&Person{})
	tag, err := obj.Field("HahaStreet").Tag("invalid")
	assert.NotNil(t, err)
	assert.Equal(t, "Invalid field HahaStreet", err.Error())
	assert.Equal(t, len(tag), 0)
}

func TestValidTag(t *testing.T) {
	obj := New(&Person{})
	tag, err := obj.Field("Street").Tag("tag")
	assert.Nil(t, err)
	assert.Equal(t, tag, "be")
}

func TestValidTags(t *testing.T) {
	obj := New(&Person{})

	tags, err := obj.Field("Street").TagExpanded("tag")
	assert.Nil(t, err)
	assert.Equal(t, tags, []string{"be"})

	tags2, err := obj.Field("Street").TagExpanded("tag2")
	assert.Nil(t, err)
	assert.Equal(t, tags2, []string{"1", "2", "3"})
}

func TestAllTags(t *testing.T) {
	obj := New(&Person{})

	tags, err := obj.Field("Street").Tags()
	assert.Nil(t, err)
	assert.Equal(t, len(tags), 2)
	assert.Equal(t, tags["tag"], "be")
	assert.Equal(t, tags["tag2"], "1,2,3")
}

func TestNewFromType(t *testing.T) {
	obj1 := NewFromType(reflect.TypeOf(Person{}))
	obj2 := New(&Person{})

	assert.Equal(t, obj1.objType.String(), obj2.objType.String())
	assert.Equal(t, obj1.objKind.String(), obj2.objKind.String())
	assert.Equal(t, obj1.underlyingType.String(), obj2.underlyingType.String())
}

func TestAnonymousFields(t *testing.T) {
	obj := New(&Person{})

	assert.True(t, obj.Field("Address").IsAnonymous())
	assert.False(t, obj.Field("Name").IsAnonymous())
}

func TestNil(t *testing.T) {
	obj := New(nil)
	assert.Equal(t, 0, len(obj.Fields()))
	assert.Equal(t, 0, len(obj.Methods()))

	res, err := obj.Field("Aaa").Get()
	assert.Nil(t, res)
	assert.NotNil(t, err)

	err = obj.Field("Aaa").Set(1)
	assert.NotNil(t, err)
}

func TestNilType(t *testing.T) {
	obj := NewFromType(nil)
	assert.Equal(t, 0, len(obj.Fields()))
	assert.Equal(t, 0, len(obj.Methods()))

	res, err := obj.Field("Aaa").Get()
	assert.Nil(t, res)
	assert.NotNil(t, err)

	err = obj.Field("Aaa").Set(1)
	assert.NotNil(t, err)

	_, err = obj.Method("aaa").Call("bu")
	assert.NotNil(t, err)
}

func TestStringObj(t *testing.T) {
	obj := New("")
	assert.Equal(t, 0, len(obj.Fields()))
	assert.Equal(t, 0, len(obj.Methods()))

	res, err := obj.Field("Aaa").Get()
	assert.Nil(t, res)
	assert.NotNil(t, err)

	err = obj.Field("Aaa").Set(1)
	assert.NotNil(t, err)

	_, err = obj.Method("aaa").Call("bu")
	assert.NotNil(t, err)
}

type TestWithInnerStruct struct {
	Aaa string
	Bbb struct {
		Ccc int
		Ddd float64
	}
}

func TestInnerStruct(t *testing.T) {
	obj := New(TestWithInnerStruct{})
	fields := obj.Fields()
	assert.Equal(t, 2, len(fields))

	assert.Equal(t, "Aaa", fields[0].Name())
	assert.Equal(t, "string", fields[0].Type().String())
	assert.Equal(t, "string", fields[0].Kind().String())

	assert.Equal(t, "Bbb", fields[1].Name())
	assert.Equal(t, "struct { Ccc int; Ddd float64 }", fields[1].Type().String())
	assert.Equal(t, "struct", fields[1].Kind().String())

	// This is not an anonymous struct, so fields are always the same:
	assert.Equal(t, 2, len(obj.FieldsAll()))
	assert.Equal(t, 2, len(obj.FieldsFlattened()))
}

func TestExportedUnexported(t *testing.T) {
	obj := New(&tmp.TestStruct{})
	assert.Equal(t, "_", obj.Fields()[0].Name())
	assert.False(t, obj.Fields()[0].IsExported())

	assert.Equal(t, "Exported", obj.Fields()[1].Name())
	assert.True(t, obj.Fields()[1].IsExported())

	assert.Equal(t, "unexported", obj.Fields()[2].Name())
	assert.False(t, obj.Fields()[2].IsExported())

	err := obj.Field("Exported").Set("aaa")
	assert.Nil(t, err)

	err = obj.Field("unexported").Set(1777)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "not settable")

	value, err := obj.Field("unexported").Get()
	assert.NotNil(t, err)
	assert.Nil(t, value)
	assert.Contains(t, err.Error(), "Cannot read unexported field")

	// But tags on unexported fields are still readable:
	{
		tags, err := obj.Field("_").Tags()
		assert.Nil(t, err)
		assert.Equal(t, 1, len(tags))
		assert.Equal(t, "ba", tags["bu"])
	}
	{
		tags, err := obj.Field("unexported").Tags()
		assert.Nil(t, err)
		assert.Equal(t, 2, len(tags))
	}
}

func TestArray(t *testing.T) {
	arr := []int{2, 3, 4}
	obj := New(arr)
	assert.Equal(t, 0, len(obj.FieldsAll()))
	assert.False(t, obj.IsArray())
	assert.True(t, obj.IsSlice())
	assert.False(t, obj.IsMap())

	l, err := obj.Len()
	assert.Nil(t, err)
	assert.Equal(t, 3, l)

	element, err := obj.GetElement(0)
	assert.Nil(t, err)
	assert.Equal(t, 2, element)
}

func TestStringAsArray(t *testing.T) {
	arr := "qwe"
	obj := New(arr)
	assert.Equal(t, 0, len(obj.FieldsAll()))
	assert.False(t, obj.IsArray())
	assert.True(t, obj.IsSlice())
	assert.False(t, obj.IsMap())

	l, err := obj.Len()
	assert.Nil(t, err)
	assert.Equal(t, 3, l)

	element, err := obj.GetElement(0)
	assert.Nil(t, err)
	assert.Equal(t, 2, element)
}

func TestMap(t *testing.T) {
	m := map[string]int{
		"a":  1,
		"bb": 17,
	}

	obj := New(m)
	assert.Equal(t, 2, len(obj.FieldsAll()))
	assert.False(t, obj.IsArray())
	assert.False(t, obj.IsSlice())
	assert.True(t, obj.IsMap())

	elem, err := obj.GetElement("a")
	assert.Nil(t, err)
	assert.Equal(t, 1, elem)
}

func TestFunction(t *testing.T) {
	// TODO
	f := func(a, b int) int { return a + b }
	obj := New(f)
	assert.Equal(t, 0, len(obj.FieldsAll()))

	assert.True(t, obj.IsFunction())

	res, err := obj.Call(1, 2)
	assert.Nil(t, err)
	assert.Equal(t, 3, res)
}
