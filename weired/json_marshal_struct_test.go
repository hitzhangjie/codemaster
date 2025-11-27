package weired

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// MyType 嵌入 time.Time 的结构体（匿名嵌入）
// 由于 time.Time 实现了 json.Marshaler 接口，嵌入它的结构体也会实现该接口
// 这会导致 json.Marshal 时只序列化 time.Time 的内容，其他字段会丢失
type MyType struct {
	ID        string
	time.Time // 嵌入字段（匿名嵌入）
}

// MyTypeWithPointer 使用指针嵌入 time.Time, 这种方式也不能避免接口提升的问题
type MyTypeWithPointer struct {
	ID         string
	*time.Time // 使用指针嵌入，也会导致接口提升 (看来只要是匿名嵌入，都会导致接口提升)
}

// MyTypeNormal 不嵌入，而是作为普通字段
type MyTypeNormal struct {
	ID   string
	Time time.Time // 普通字段（命名字段），不会导致接口提升
}

func TestJSONMarshalWithEmbeddedTime(t *testing.T) {
	now := time.Date(2025, 11, 27, 10, 0, 0, 0, time.Local)

	// 测试嵌入 time.Time 的情况
	myType := MyType{
		ID:   "1234",
		Time: now, // 嵌入字段通过类型名访问
	}
	data, err := json.Marshal(myType)
	require.Nil(t, err)
	t.Logf("MyType (embedded time.Time) JSON: %s", string(data))

	// 只marshal now
	want, err := now.MarshalJSON()
	require.Nil(t, err)

	// 因为myType.Time是匿名嵌入的，Time的json.Marshaler接口会被提升，所以只会marshal now，
	// 很多开发者预期的结果是，应该包含ID字段吧，实际上不会。
	require.Equal(t, string(want), string(data))
}

func TestJSONMarshalWithPointerTime(t *testing.T) {
	now := time.Date(2025, 11, 27, 10, 0, 0, 0, time.Local)

	// 测试使用指针嵌入的情况
	myType := MyTypeWithPointer{
		ID:   "1234",
		Time: &now, // 指针嵌入字段
	}
	data, err := json.Marshal(myType)
	require.Nil(t, err)
	t.Logf("MyTypeWithPointer (pointer to time.Time) JSON: %s", string(data))

	want, err := now.MarshalJSON()
	require.Nil(t, err)

	// 因为myType.Time是指针嵌入的，Time的json.Marshaler接口也是会被提升的，所以也不会包含ID字段，
	require.Equal(t, string(want), string(data))
}

func TestJSONMarshalWithNormalTime(t *testing.T) {
	now := time.Date(2025, 11, 27, 10, 0, 0, 0, time.Local)

	// 测试普通字段的情况
	myType := MyTypeNormal{
		ID:   "1234",
		Time: now,
	}
	data, err := json.Marshal(myType)
	require.Nil(t, err)
	t.Logf("MyTypeNormal (normal time.Time field) JSON: %s", string(data))

	want, err := now.MarshalJSON()
	require.Nil(t, err)

	require.NotEqual(t, string(want), string(data))
}

func TestJSONMarshalComparison(t *testing.T) {
	now := time.Now()

	// 嵌入 time.Time
	embedded := MyType{
		ID: "1234",
	}
	embedded.Time = now

	// 普通字段
	normal := MyTypeNormal{
		ID:   "1234",
		Time: now,
	}

	embeddedData, _ := json.Marshal(embedded)
	normalData, _ := json.Marshal(normal)

	t.Logf("Embedded time.Time: %s", string(embeddedData))
	t.Logf("Normal time.Time field: %s", string(normalData))

	// 验证两者不同
	if string(embeddedData) == string(normalData) {
		t.Errorf("Expected different results, but they are the same")
	}

	// 验证嵌入的情况确实丢失了 ID 字段
	var embeddedResult map[string]interface{}
	json.Unmarshal(embeddedData, &embeddedResult)

	var normalResult map[string]interface{}
	json.Unmarshal(normalData, &normalResult)

	if _, exists := embeddedResult["ID"]; exists {
		t.Errorf("Embedded struct should not have ID field, but it does")
	}

	if _, exists := normalResult["ID"]; !exists {
		t.Errorf("Normal struct should have ID field, but it doesn't")
	}
}
