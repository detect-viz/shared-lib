## SpecStory

## Fixing Undefined Error in Go Code (2025/2/13 下午3:13:15)

_**User**_

For the code present, we get this error:
```
undefined: common.Value
```
How can I resolve this? If you propose a fix, please make it concise.

---

_**Assistant**_

看起來是 `common.Value` 常量未在 `common` 包中定義。我們需要在 common 包中添加這個常量。

以下是修復建議：

```go:parser/logman/parser.go
// ... existing code ...
	"shared-lib/parser/common"
// ... existing code ...
```

您需要在 `shared-lib/parser/common/constants.go`（或類似的文件）中添加這個常量定義：

```go:shared-lib/parser/common/constants.go
// ... existing code ...
const (
    Value = "value"
)
// ... existing code ...
```

這樣就可以在解析器中正常使用 `common.Value` 了。如果您已經有一個常量文件，只需添加 `Value` 常量即可。

---



---

