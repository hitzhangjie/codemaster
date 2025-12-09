package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Event struct {
	Time string `json:"time"`
}

// 用于测试 time.Time 在格式化和解析时 layout 带不带时区的表现差异
//
// Parse 在处理时区时确实很容易让人产生误解：
//
//  1. 无时区标记：默认解析为 UTC（不是本地时区），这与 Format 时的行为不同。
//  2. 有时区偏移（如 -0700）：只有当本地时区配置的偏移值和字符串一致时，才会落在本地 Location，否则会新建一个“虚构”时区。
//     如果时区配置错误，比如+0800，这个+0800也会被追加到格式化时间字符串中，不容易察觉到此配置错误，parse时才会暴漏问题。
//  3. RFC3339/RFC3339Nano：这是一个常量格式化串，建议使用这两个常量，减少配置错误的机会。
//  4. 有缩写（如 MST/CST/GMT）：格式化时，MST 占位符会被替换为实际时区缩写。
//     - 解析时，如果字符串中的缩写与本地时区的缩写匹配，则使用本地 Location；
//     - 如果与本地已知的某个时区缩写匹配，则使用该时区，比如Asia/Shanghai，则使用Asia/Shanghai时区及偏移量。
//     - 如果不匹配，则创建一个偏移量为 0 的FixedZone（这可能导致时间错误！）。
//     warning：某些缩写有歧义（如 CST 可能指中国标准时间 UTC+8 或美国中部标准时间 UTC-6），建议使用明确的时区偏移量。
//
// 总之，Parse 行为确实不直观，实际用时如需明确时区建议用 ParseInLocation。详见官方注释和你给的测试用例行为。
func Test_formatAndParseTimezoneOffset(t *testing.T) {
	// time.Time.location = time.Local
	now := time.Now()

	// note:
	// - layout没有timezone，format时使用time.Time.location (time.Local) 作为时区，也就不会有额外的时区偏移量
	// - layout有timezone，parse时间字符串时，不是默认用time.Local，而是用UTC时区，
	//
	// mistake: 如果使用同一个没有timezone的layout格式化，然后再parse，此时相当于丢失了时区偏移量
	t.Run("(badcase) without timezone: format and parse, the result loses timezone offset info", func(t *testing.T) {
		layoutNoTZ := "2006-01-02 15:04:05.999999999"

		// Format: 默认使用 time.Local 计算时间偏移量，然后转换为字符串
		formatted := now.Format(layoutNoTZ)
		formattedLoc := now.In(time.Local).Format(layoutNoTZ)
		assert.Equal(t, formatted, formattedLoc, "Format 和 Format/In(Local) 结果应一致")

		// Marshal/Unmarshal 测试
		tm, _ := json.Marshal(Event{Time: formatted})
		var evt Event
		_ = json.Unmarshal(tm, &evt)

		// Parse: 由于layout里面没有时区信息，解析为unix时间戳时就不会有额外的时区偏移量
		parsed, err := time.Parse(layoutNoTZ, evt.Time)
		assert.NoError(t, err)

		// 此时如果基于unixnano进行对比，发现是不相同的，因为parse时layout里丢失了时区偏移量
		assert.Equal(t, "UTC", parsed.Location().String(), "layout不带时区parse出来总是UTC")
		assert.NotEqual(t, now.UnixNano(), parsed.UnixNano(), "unixnano 应不同，因为丢失了时区便宜量")
	})

	// note:
	// - layout有正确的timezone配置，注意是'07'，就跟2006表示年，01表示月，02表示日，15表示小时……一样的，07表示UTC时区偏移量
	// - layout有正确的timezone配置，格式化时就会追加UTC时区偏移量到时间字符串中
	// - layout中没有正确的timezone配置，格式化时会将其他字符串原样追加回去，比如误认为使用+0800表示时区偏移量，是错误的
	//   - 格式化时可能察觉不到，因为time.Local确实是+0800时区
	//   - 解析时会察觉到，因为parse时layout里丢失了时区偏移量，因为不认+0800，只认-0700
	// - layout中有正确的timezone配置，解析时发现时间字符串中的UTC时区偏移量和当前time.Local时区偏移量是否一致：
	//    - 一致：此时time.Time.location=time.Local
	//    - 不一致：此时time.Time.location=UTC
	t.Run("(goodcase) with timezone -0700: format and parse, the result keeps timezone offset info", func(t *testing.T) {
		layoutWithTZ := "2006-01-02 15:04:05.999999999 -0700"
		// now当前时区是time.Local，按照layout格式化时要求将UTC时区偏移量追加到时间字符串中
		// 如time.Local=Asia/Shanghai，则格式化后时间字符串为：YYYY-MM-DD hh:mm:ss.xxxxxxxx +0800
		formatted := now.Format(layoutWithTZ)
		fmt.Printf("formatted: %v\n", formatted)

		// 因为layout中有正确的timezone配置，解析时发现时间字符串中的UTC时区偏移量和当前time.Local时区偏移量一致，
		// 所以time.Time.location=time.Local
		parsed, _ := time.Parse(layoutWithTZ, formatted)
		assert.Equal(t, "Local", parsed.Location().String(), "time.Time.location=time.Local，因为layout timezone配置正确，且时间字符串中的UTC时区偏移量和当前time.Local时区偏移量一致")
		assert.Equal(t, now.UnixNano(), parsed.UnixNano(), "时间点应完全一致")
	})

	t.Run("(badcase) with wrong timezone +0800: format and parse, the result loses timezone offset info", func(t *testing.T) {
		layoutWithTZ := "2006-01-02 15:04:05.999999999 +0800"
		// now当前时区是time.Local，按照layout格式化时要求将UTC时区偏移量追加到时间字符串中，
		// 但是layout中有错误的timezone配置，+0800，所以格式化后时间字符串为：YYYY-MM-DD hh:mm:ss.xxxxxxxx +0800，视觉上看上去没问题，因为+0800会简单追加。
		// ps: 要等到parse时才会暴漏问题。
		formatted := now.Format(layoutWithTZ)
		fmt.Printf("formatted: %v\n", formatted)

		location, _ := time.LoadLocation("Asia/Shanghai")
		formattedInShanghai := now.In(location).Format(layoutWithTZ[:len(layoutWithTZ)-5] + "-0700")
		fmt.Printf("formattedInShanghai: %v\n", formattedInShanghai)

		assert.Equal(t, formattedInShanghai, formatted, "格式化后时间字符串应完全一致")

		// 因为layout中没有正确的timezone配置，+0800是错误的，必须用-0700来解析timezone，所以time.Time.location=time.UTC，
		// 此时也就相当于丢失了原来的时区偏移量信息（UTC +8小时）
		parsed, _ := time.Parse(layoutWithTZ, formatted)
		assert.Equal(t, "UTC", parsed.Location().String(), "layout中没有正确的timezone配置，+0800是错误的，必须用-0700来解析timezone，所以time.Time.location=time.UTC")
		assert.NotEqual(t, now.UnixNano(), parsed.UnixNano(), "layout没有正确的timezone配置，+0800是错误的，时区偏移量丢失，时间点应该不一致")
	})

	t.Run("(goodcase) with RFC3339/RFC3339Nano, suggested to use", func(t *testing.T) {
		// 这里 layout 字符串与时区需保持一致，并明确写 +0800 和 CST 代表中国标准时间。
		layoutWithTZ := time.RFC3339Nano
		formatted := now.Format(layoutWithTZ)
		fmt.Printf("formatted: %v\n", formatted)

		// Parse: time.Parse 的行为说明，注意看文档说明以及实现，只有RFC3339会使用layout中指定的时区信息，
		// 其他情况下会使用UTC
		parsed, _ := time.Parse(layoutWithTZ, formatted)
		assert.Equal(t, "Local", parsed.Location().String(), "RFC3339Nano 解析后时区偏移量与Local相同，所以为Local")
		assert.Equal(t, now.UnixNano(), parsed.UnixNano(), "解析后的时间点应与原时间完全一致")
	})

	t.Run("(goodcase) with RFC3339/RFC3339Nano and fixed timezone, suggested to use", func(t *testing.T) {
		// 这里 layout 字符串与时区需保持一致，并明确写 +0800 和 CST 代表中国标准时间。
		layoutWithTZ := time.RFC3339Nano
		formatted := now.Format(layoutWithTZ)
		fmt.Printf("formatted: %v\n", formatted)

		// Parse: time.Parse 的行为说明，注意看文档说明以及实现，只有RFC3339会使用layout中指定的时区信息，
		// 其他情况下会使用UTC
		parsed, _ := time.Parse(layoutWithTZ, formatted)
		assert.Equal(t, "Local", parsed.Location().String(), "RFC3339Nano 解析后时区偏移量与Local相同，所以为Local")
		assert.Equal(t, now.UnixNano(), parsed.UnixNano(), "解析后的时间点应与原时间完全一致")

		// ok, time.Local是东八区，这里测试用例设计没那么严格有效，知道就好了。
		tzInEast9 := time.FixedZone("东9区", 9*3600)
		formattedInEast9 := now.In(tzInEast9).Format(layoutWithTZ)
		fmt.Printf("formattedInEast9: %v\n", formattedInEast9)
		assert.NotEqual(t, formattedInEast9, formatted, "格式化后时间字符串不应该与原时间字符串相同，差1个小时偏移量")

		parsedInEast9, _ := time.Parse(layoutWithTZ, formattedInEast9)

		// 为什么parse出来，Location不是UTC呢？UTC指的是UTC 0时区，即格林尼治时间。
		// 因为时间字符串包含了时区信息+09:00，所以Go会为其创建一个匿名FixedZone，而不是UTC。
		// 这个匿名FixedZone的name是""，偏移量为+09:00。
		// 只有完全没有时区信息时，才会默认用UTC。
		assert.Equal(t, "", parsedInEast9.Location().String(), "时间字符串含有偏移量(+09:00)但与Local不一致，Go将为其创建一个匿名FixedZone")
		assert.Equal(t, now.UnixNano(), parsedInEast9.UnixNano(), "解析后的时间点应与原时间完全一致")
	})

	// 时区缩写（如 CST、MST、GMT）的作用说明：
	// 1. 在 layout 格式字符串中，MST 是占位符，代表时区缩写应该插入的位置。
	// 2. Format 时：MST 会被当前时间所属 Location 的时区缩写实际值（如 CST、PST、GMT 等）替换。
	// 3. Parse 时：
	//    - 如果被解析字符串的缩写与本地 Location（time.Local）的某时区缩写相符，则解析结果的 Location 也为本地 Location。
	//    - 如果缩写在本地时区数据库中的任意已知时区内有定义，则会使用该缩写和其对应的固定偏移量（即创建有名的 FixedZone，名字是缩写，偏移量是该缩写的标准定义值）。
	//    - 如果缩写既不匹配本地 Location，也不在本地时区数据库中定义，则会创建一个匿名 FixedZone，其偏移量为 0（即 UTC）。
	// 4. 注意：某些缩写（比如 CST）在全球范围内并非唯一，可能表示不同的时区（如中国标准时间 UTC+8，美国中部标准时间 UTC-6 等），因此可能存在歧义。
	// 5. 并不是每个时区都对应唯一的缩写。有的时区可能没有缩写，或者某些缩写会被多个不同时区共用，因此仅靠缩写无法唯一确定时区及其偏移量。
	t.Run("timezone abbreviation (MST/CST/GMT): format and parse behavior", func(t *testing.T) {
		now := time.Now()

		// 格式化时：MST 会被替换为实际的时区缩写
		// 如果 time.Local 是 Asia/Shanghai，则输出 CST
		// 如果 time.Local 是 America/New_York，则输出 EST/EDT（取决于是否夏令时）
		layoutWithAbbr := "2006-01-02 15:04:05 MST"
		formatted := now.Format(layoutWithAbbr)
		fmt.Printf("formatted with MST: %v\n", formatted)
		fmt.Printf("current location: %v, abbreviation: %v\n", time.Local.String(), now.Format("MST"))

		// 解析时：如果字符串中的缩写与本地时区匹配，使用本地 Location
		parsed, err := time.Parse(layoutWithAbbr, formatted)
		assert.NoError(t, err)
		fmt.Printf("parsed location: %v\n", parsed.Location().String())

		// 测试：如果本地时区定义了该缩写，则使用本地 Location
		// 注意：这取决于系统时区配置，可能因环境而异
		localAbbr := now.Format("MST")
		if localAbbr != "" {
			// 如果本地时区有缩写且匹配，应该使用本地 Location
			// 但实际行为可能因 Go 版本和时区数据而异
			fmt.Printf("Local abbreviation matches, location should be Local or a FixedZone\n")
		}

		// 测试：解析一个不匹配本地时区的缩写（如 GMT）
		// 如果本地时区不是 GMT，解析时会创建一个偏移量为 0 的虚构时区
		gmtTimeStr := "2024-01-01 12:00:00 GMT"
		parsedGMT, err := time.Parse(layoutWithAbbr, gmtTimeStr)
		assert.NoError(t, err)
		fmt.Printf("parsed GMT location: %v, offset: %v\n", parsedGMT.Location().String(), parsedGMT.Format("-0700"))

		// 测试：解析 CST（有歧义的缩写）
		// CST 可能指中国标准时间（UTC+8）或美国中部标准时间（UTC-6）
		cstTimeStr := "2024-01-01 12:00:00 CST"
		parsedCST, err := time.Parse(layoutWithAbbr, cstTimeStr)
		assert.NoError(t, err)
		fmt.Printf("parsed CST location: %v, offset: %v\n", parsedCST.Location().String(), parsedCST.Format("-0700"))
		fmt.Printf("WARNING: CST is ambiguous! It could be China Standard Time (UTC+8) or Central Standard Time (UTC-6)\n")

		// 对比：使用明确的时区偏移量（推荐方式）
		layoutWithOffset := "2006-01-02 15:04:05.999999999 -0700"
		formattedWithOffset := now.Format(layoutWithOffset)
		parsedWithOffset, err := time.Parse(layoutWithOffset, formattedWithOffset)
		assert.NoError(t, err)
		fmt.Printf("formatted with offset: %v\n", formattedWithOffset)
		fmt.Printf("parsed with offset location: %v\n", parsedWithOffset.Location().String())
		assert.Equal(t, now.UnixNano(), parsedWithOffset.UnixNano(), "使用偏移量时，时间点应完全一致")
	})
}

func Test_formatTimeZone(t *testing.T) {
	location := time.FixedZone("Asia/Shanghai", 8*3600)
	now := time.Unix(1734583230, 0)
	println("now.Unix:", now.Unix())
	println("now.Location:", now.Location().String())
	println("now.In(Asia/Shanghai):", now.In(location).Location().String())
	println("now.UTC():", now.UTC().Location().String())

	// 1. 错误姿势：layout写“UTC +0000” 实际只影响输出的内容，不影响时间格式化和偏移
	//
	// 实际效果：layout 里 "UTC" 是普通字符串，"+0000" 也是普通字符串，不会自动换成真实时区信息
	// 只有 -0700 之类的 layout 通配符才会被替换为时区偏移
	layoutWrong := "2006-01-02 15:04:05 UTC +0000"
	println("[1] layoutWrong(UTC +0000):")
	println("   now.         :", now.Format(layoutWrong))
	println("   now.Local()  :", now.Local().Format(layoutWrong))
	println("   now.In(8区)  :", now.In(location).Format(layoutWrong))
	println("   now.In(UTC)  :", now.In(time.UTC).Format(layoutWrong))
	println("   now.UTC()    :", now.UTC().Format(layoutWrong))

	// 2. 正确姿势：layout 用 -0700（时区偏移格式化必选）
	layoutCorrect := "2006-01-02 15:04:05 -0700"
	println("\n[2] layoutCorrect(-0700):")
	println("   now.         :", now.Format(layoutCorrect))
	println("   now.Local()  :", now.Local().Format(layoutCorrect))
	println("   now.In(8区)  :", now.In(location).Format(layoutCorrect))
	println("   now.In(UTC)  :", now.In(time.UTC).Format(layoutCorrect))
	println("   now.UTC()    :", now.UTC().Format(layoutCorrect))

	// 3. 如果 layout 字符串里包含XXX或者非time包格式化、解析认定的一些字面值，如-0700 or MST等，
	// 就只会原样输出layout字面，不会根据时间本身的Location决定是否输出。
	layoutCorrectWithUTCTerm := "2006-01-02 15:04:05 XXX -0700"
	println("\n[3] layout 里含 'XXX' 与解析无关的字符串，会原样输出：")
	println("   now.Local().Format(layoutWrong)    :", now.Local().Format(layoutCorrectWithUTCTerm))
	println("   now.UTC().Format(layoutWrong)      :", now.UTC().Format(layoutCorrectWithUTCTerm))
}
