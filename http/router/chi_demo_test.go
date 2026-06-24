package router_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// =============================================================================
// chi 路由参数解析演示
// =============================================================================
//
// chi 的 RESTful 路由参数通过 {paramName} 或 {paramName:regex} 定义在 URL 模式中。
// 参数值存储在 chi.Context.URLParams 中，通过 chi.URLParam(r, key) 获取。
//
// 核心数据结构：
//
//   type Context struct {
//       URLParams    RouteParams   // 所有已匹配的参数（跨子路由累积）
//       routeParams  RouteParams   // 当前子路由匹配的参数（未导出）
//       RoutePatterns []string     // 已匹配的路由模式栈
//       // ...
//   }
//
//   type RouteParams struct {
//       Keys   []string
//       Values []string
//   }
//
// URLParams 的查找是从后往前遍历（倒序），这意味着：
// - 最内层子路由的参数优先级最高
// - 如果外层和内层有同名参数，内层的值胜出
//
// 这与 chi 的递归路由匹配机制一致：请求进入时，从根 router 开始，逐层
// 匹配子路由，每匹配一层就将该层的参数追加到 URLParams 末尾。

// TestBasicURLParam 演示最基本的 RESTful 路径参数定义与提取。
func TestBasicURLParam(t *testing.T) {
	r := chi.NewRouter()

	// 用 {paramName} 语法定义路径参数
	r.Get("/projects/{projectID}", func(w http.ResponseWriter, r *http.Request) {
		// chi.URLParam 从请求上下文中提取参数值
		id := chi.URLParam(r, "projectID")
		w.Write([]byte("project:" + id))
	})

	// 发送请求并验证
	req := httptest.NewRequest("GET", "/projects/42", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Body.String() != "project:42" {
		t.Fatalf("expected 'project:42', got %q", rec.Body.String())
	}
}

// TestMultipleURLParams 演示一个路径中定义多个参数。
func TestMultipleURLParams(t *testing.T) {
	r := chi.NewRouter()

	// 多个参数用多个 {name} 定义
	r.Get("/repos/{owner}/{repo}/issues/{issueID}", func(w http.ResponseWriter, r *http.Request) {
		owner := chi.URLParam(r, "owner")
		repo := chi.URLParam(r, "repo")
		issueID := chi.URLParam(r, "issueID")
		w.Write([]byte(owner + "/" + repo + "#" + issueID))
	})

	tests := []struct {
		path string
		want string
	}{
		{"/repos/alice/myrepo/issues/101", "alice/myrepo#101"},
		{"/repos/bob/another-repo/issues/42", "bob/another-repo#42"},
	}

	for _, tc := range tests {
		req := httptest.NewRequest("GET", tc.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Body.String() != tc.want {
			t.Errorf("%s: got %q, want %q", tc.path, rec.Body.String(), tc.want)
		}
	}
}

// TestRegexURLParam 演示带正则约束的路径参数。
//
// 注意：{param:regex} 中的正则永远不能匹配 '/'，只能匹配单个 path segment。
// 若需要匹配含 '/' 的多段路径，应使用通配符 /*（见 TestWildcardRoute）。
func TestRegexURLParam(t *testing.T) {
	r := chi.NewRouter()

	// {param:regex} 限制参数格式：仅允许字母、数字、点、下划线、连字符，
	// NOTE: 也可以用*，但是还是不能跨越/ 。
	r.Get("/files/{fileName:*}", func(w http.ResponseWriter, r *http.Request) {
		fileName := chi.URLParam(r, "fileName")
		w.Write([]byte("file:" + fileName))
	})

	tests := []struct {
		path string
		want string
		code int
	}{
		{"/files/README.md", "file:README.md", 200},
		{"/files/main.go", "file:main.go", 200},
		{"/files/my-file_v2.txt", "file:my-file_v2.txt", 200},
		{"/files/src/main.go", "", 404}, // 含 '/' 的路径无法被 {param:regex} 匹配
	}

	for _, tc := range tests {
		req := httptest.NewRequest("GET", tc.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Code != tc.code {
			t.Errorf("%s: status %d, want %d", tc.path, rec.Code, tc.code)
		}
		if tc.code == http.StatusOK && rec.Body.String() != tc.want {
			t.Errorf("%s: body %q, want %q", tc.path, rec.Body.String(), tc.want)
		}
	}
}

// TestSubrouterParamOverriding 演示子路由参数倒序查找的优先级规则。
//
// 这是 chi.URLParam 倒序查找的关键场景：
// 当外层路由和内层子路由定义了同名参数时，内层的值胜出。
func TestSubrouterParamOverriding(t *testing.T) {
	r := chi.NewRouter()

	// 外层路由：/orgs/{name}/
	r.Route("/orgs/{name}", func(r chi.Router) {
		// 内层路由：/{name}/settings
		// 注意：内层也用了 {name}，与外层同名
		r.Get("/{name}/settings", func(w http.ResponseWriter, r *http.Request) {
			// URLParam 倒序查找，所以返回的是内层的 {name} 值
			name := chi.URLParam(r, "name")
			w.Write([]byte("settings for:" + name))
		})

		// 这个路由引用的是外层的 {name}
		r.Get("/info", func(w http.ResponseWriter, r *http.Request) {
			name := chi.URLParam(r, "name")
			w.Write([]byte("org info:" + name))
		})
	})

	// 请求 /orgs/acme-corp/security-team/settings
	// URLParams 栈：
	//   Keys: ["name", "name"]
	//   Values: ["acme-corp", "security-team"]
	// chi.URLParam(r, "name") 倒序查找 → 先匹配到 "security-team"
	req := httptest.NewRequest("GET", "/orgs/acme-corp/security-team/settings", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if got := rec.Body.String(); got != "settings for:security-team" {
		t.Errorf("expected inner param win, got %q", got)
	}

	// 请求 /orgs/acme-corp/info — 只有外层的 {name}
	req2 := httptest.NewRequest("GET", "/orgs/acme-corp/info", nil)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if got := rec2.Body.String(); got != "org info:acme-corp" {
		t.Errorf("expected outer param, got %q", got)
	}
}

// TestChiURLParamInternal 用 httptest 直接观察 chi 内部 Context 的 URLParams 变化。
//
// 通过中间件，我们可以在路由匹配前后查看 URLParams 的累积情况。
func TestChiURLParamInternal(t *testing.T) {
	r := chi.NewRouter()

	// 中间件：在路由匹配后打印 URLParams 状态
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 中间件执行时，路由已经匹配完成，URLParams 已填充
			ctx := chi.RouteContext(r.Context())
			t.Logf("After route match:")
			t.Logf("  URLPattern: %s", ctx.RoutePattern())
			t.Logf("  URLParams.Keys:   %v", ctx.URLParams.Keys)
			t.Logf("  URLParams.Values: %v", ctx.URLParams.Values)
			next.ServeHTTP(w, r)
		})
	})

	// 定义一个嵌套路由结构
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/users/{userID}/posts/{postID}", func(w http.ResponseWriter, r *http.Request) {
			rctx := chi.RouteContext(r.Context())

			// 在 handler 中可以观察到完整的参数信息
			t.Logf("In handler:")
			t.Logf("  RoutePatterns:  %v", rctx.RoutePatterns)
			t.Logf("  URLParams.Keys: %v, Values: %v", rctx.URLParams.Keys, rctx.URLParams.Values)

			userID := chi.URLParam(r, "userID")
			postID := chi.URLParam(r, "postID")
			w.Write([]byte("user=" + userID + " post=" + postID))
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/users/7/posts/99", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	want := "user=7 post=99"
	if rec.Body.String() != want {
		t.Fatalf("got %q, want %q", rec.Body.String(), want)
	}
}

// TestWildcardRoute 演示 * 通配符路由。
func TestWildcardRoute(t *testing.T) {
	r := chi.NewRouter()

	// /* 匹配所有剩余路径（类似文件系统的 glob）
	r.Get("/static/*", func(w http.ResponseWriter, r *http.Request) {
		// 通配符部分通过 "*" 这个 key 获取
		filePath := chi.URLParam(r, "*")
		w.Write([]byte("serving:" + filePath))
	})

	tests := []struct {
		path string
		want string
	}{
		{"/static/logo.png", "serving:logo.png"},
		{"/static/css/main.css", "serving:css/main.css"},
		{"/static/js/vendor/lib.js", "serving:js/vendor/lib.js"},
	}

	for _, tc := range tests {
		req := httptest.NewRequest("GET", tc.path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)

		if rec.Body.String() != tc.want {
			t.Errorf("%s: got %q, want %q", tc.path, rec.Body.String(), tc.want)
		}
	}
}

// TestUnknownParamReturnsEmpty 演示访问不存在的参数名时返回空字符串。
func TestUnknownParamReturnsEmpty(t *testing.T) {
	r := chi.NewRouter()

	r.Get("/items/{itemID}", func(w http.ResponseWriter, r *http.Request) {
		// 请求一个不存在的参数名
		val := chi.URLParam(r, "nonexistent")
		if val != "" {
			t.Errorf("expected empty string for unknown param, got %q", val)
		}
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest("GET", "/items/123", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
}
