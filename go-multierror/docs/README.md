我们先来定义一个自定义错误 `Error` 的结构体

这个结构体有两个字段

- `Errors` 存储错误信息的切片
- `ErrorFormat` 存储错误格式化函数

```go
type Error struct {
  Errors      []error
  ErrorFormat ErrorFormatFunc
}

func (e *Error) Error() string {
  fn := e.ErrorFormat
  if fn == nil {
    fn = ListFormatFunc
  }

  return fn(e.Errors)
}
```

这个自定义的结构体需要实现 `error` 接口，才是一个 `Error` 类型的错误

`Error` 方法主要的作用返回错误信息，这里内部调用 `ErrorFormat` 函数，格式化信息

如果没有设置 `ErrorFormat` 函数，那么默认使用 `ListFormatFunc` 函数

我们来看一下 `ListFormatFunc` 函数，接收一个错误切片，返回一个格式化的错误信息

```go
func ListFormatFunc(es []error) string {
  if len(es) == 1 {
    return fmt.Sprintf("1 error occurred:\n\t* %s\n\n", es[0])
  }

  points := make([]string, len(es))
  for i, err := range es {
    points[i] = fmt.Sprintf("* %s", err)
  }

  return fmt.Sprintf(
    "%d errors occurred:\n\t%s\n\n",
    len(es), strings.Join(points, "\n\t"))
}
```

我们来看对 `ListFormatFunc` 函数的 `2` 个测试用例：

```go
func TestListFormatFuncSingle(t *testing.T) {
  expected := `1 error occurred:
  * foo

`

  errors := []error{
    errors.New("foo"),
  }

  actual := ListFormatFunc(errors)
  if actual != expected {
    t.Fatalf("bad: %#v", actual)
  }
}
```

```go
func TestListFormatFuncMultiple(t *testing.T) {
  expected := `2 errors occurred:
  * foo
  * bar

`

  errors := []error{
    errors.New("foo"),
    errors.New("bar"),
  }

  actual := ListFormatFunc(errors)
  if actual != expected {
    t.Fatalf("bad: %#v", actual)
  }
}
```

## Append

我们来看第一个接口 `Append`

`Append` 方法用来合并多个错误，返回多个错误合并后的错误切片

我们来看它的第一个测试用例

### TestAppend_Error

第一个测试用例 `TestAppend_Error` 传入两个原生的 `error` 类型的错误，我们来检查这两个错误是不是被合并到自定义 `Error` 上了

```go
func TestAppend_NonError(t *testing.T) {
  original := errors.New("foo")
  result := Append(original, errors.New("bar"))
  if len(result.Errors) != 2 {
    t.Fatalf("wrong len: %d", len(result.Errors))
  }
}
```

实现这个功能，`Append` 函数需要接收两个参数，一个 `error` 类型的错误，一个 `error` 类型的错误切片

在 `Append` 函数内部，把第一个参数 `error` 合并到 `errors` 切片中

然后在遍历 `errors` 的切片，将每一个 `error` 转换成自定义的 `Error` 类型

```go
func Append(err error, errs ...error) *Error {
  newErrs := make([]error, 0, len(errs)+1)
  if err != nil {
    newErrs = append(newErrs, err)
  }
  newErrs = append(newErrs, errs...)
  err1 := new(Error)

  for _, e := range newErrs {
    err1.Errors = append(err1.Errors, e)
  }
  return err1
}
```

### TestAppend_NonError_Error

我们来看第二个测试用例 `TestAppend_NonError_Error`

这个测试用例测试的是递归调用 `Append` 函数

```go
func TestAppend_NonError_Error(t *testing.T) {
  original := errors.New("foo")
  result := Append(original, Append(nil, errors.New("bar")))
  if len(result.Errors) != 2 {
    t.Fatalf("wrong len: %d", len(result.Errors))
  }
}
```

上面的代码可以满足这个测试用例

### TestAppend_NilError

我们再来看第三个测试用例 `TestAppend_NilError`，我们传入的是一个 `nil` 类型，这个测试用例也能够运行

```go
func TestAppend_NilError(t *testing.T) {
  var err error
  result := Append(err, errors.New("bar"))
  if len(result.Errors) != 1 {
    t.Fatalf("wrong len: %d", len(result.Errors))
  }
}
```

### TestAppend_NilErrorArg

我们再来看第四个测试用例 `TestAppend_NilErrorArg`

我们知道指针 `*Error` 初始值是 `nil`

那么作为第二个参数传入，就会变成 `nil` 的切片：`[nil]`

```go
func TestAppend_NilErrorArg(t *testing.T) {
  var err error
  var nilErr *Error
  result := Append(err, nilErr)
  fmt.Println("result:", result)
  if len(result.Errors) != 0 {
    t.Fatalf("wrong len: %d", len(result.Errors))
  }
}
```

那么如何实现这个功能呢？

我们可以通过断言来判断当成的参数是不是 `*Error` 类型

```go
switch err.(type) {
case *Error:
default:
}
```

正常情况走下面 `default` 分支

然后递归调用 `Append`，传入 `&Error{}` 和合并好的 `errs` 切片

再次进入 `Append` 函数时，`err` 就是 `*Error` 类型了，遍历 `errs` 将每个 `err` 转成 `Error` 类型就可以了

```go
func Append(err error, errs ...error) *Error {
  switch err.(type) {
  case *Error:
    err1 := new(Error)
    for _, e := range errs {
      err1.Errors = append(err1.Errors, e)
    }
    return err1
  default:
    newErrs := make([]error, 0, len(errs)+1)
    if err != nil {
      newErrs = append(newErrs, err)
    }
    newErrs = append(newErrs, errs...)
    return Append(&Error{}, newErrs...)
  }
}
```

但是问题还没有解决，就是 `errs` 是 `nil` 切片的问题

解决这个问题就是在遍历 `errs` 时，对每一个 `err` 进行断言，如果是 `*Error` 类型，那么就直接合并到 `e.Errors` 切片中，如果不是 `Errors` 类型，那么就直接将 `e` 合并到 `e.Errors` 中

```go
func Append(err error, errs ...error) *Error {
  switch newErr := err.(type) {
  case *Error:
    for _, e := range errs {
      switch e := e.(type) {
      case *Error:
        if e != nil {
          newErr.Errors = append(newErr.Errors, e.Errors...)
        }
      default:
        newErr.Errors = append(newErr.Errors, e)
      }
    }
    return newErr
  default:
    newErrs := make([]error, 0, len(errs)+1)
    if err != nil {
      newErrs = append(newErrs, err)
    }
    newErrs = append(newErrs, errs...)
    return Append(&Error{}, newErrs...)
  }
}
```

### TestAppend_NilErrorIfaceArg

我们再来看第六个测试用例 `TestAppend_NilErrorIfaceArg`

相比于第五个测试用例，这个测试用例传入的是两个 `nil` 类型

```go
func TestAppend_NilErrorIfaceArg(t *testing.T) {
  var err error
  var nilErr error
  result := Append(err, nilErr)
  if len(result.Errors) != 0 {
    t.Fatalf("wrong len: %d", len(result.Errors))
  }
}
```

需要在 `default` 里面加上 `e != nil` 的判断

```go
for _, e := range errs {
  switch e := e.(type) {
  case *Error:
    if e != nil {
      newErr.Errors = append(newErr.Errors, e.Errors...)
    }
  default:
    if e != nil {
      newErr.Errors = append(newErr.Errors, e)
    }
  }
}
```

### TestAppend_Error

最后一个测试用例 `TestAppend_Error` 测试的是错误信息的合并

```go
func TestAppend_Error(t *testing.T) {
  original := &Error{
    Errors: []error{errors.New("foo")},
  }

  result := Append(original, errors.New("bar"))
  if len(result.Errors) != 2 {
    t.Fatalf("wrong len: %d", len(result.Errors))
  }

  original = &Error{}
  result = Append(original, errors.New("bar"))
  fmt.Println("result", result)
  if len(result.Errors) != 1 {
    t.Fatalf("wrong len: %d", len(result.Errors))
  }

  // Test when a typed nil is passed
  var e *Error
  result = Append(e, errors.New("baz"))
  if len(result.Errors) != 1 {
    t.Fatalf("wrong len: %d", len(result.Errors))
  }

  // Test flattening
  original = &Error{
    Errors: []error{errors.New("foo")},
  }

  result = Append(original, Append(nil, errors.New("foo"), errors.New("bar")))
  if len(result.Errors) != 3 {
    t.Fatalf("wrong len: %d", len(result.Errors))
  }
}
```

`Append` 函数的测试用例全部通过了

## Flatten

我们来看第二个接口 `Flatten`

`Flatten` 方法用来展平错误，返回一个单一的 `*Error` 类型的错误

我们先来看测试用例 `TestFlatten`，在这个测试用例中，`original` 是一个嵌套的 `*Error` 类型的错误，它里面嵌套了三层 `*Error` 的结构

`Flatten` 方法的作用就是将 `original` 中的 `*Error` 展平

```go
func TestFlatten(t *testing.T) {
  original := &Error{
    Errors: []error{
      errors.New("one"),
      &Error{
        Errors: []error{
          errors.New("two"),
          &Error{
            Errors: []error{
              errors.New("three"),
            },
          },
        },
      },
    },
  }

  expected := `3 errors occurred:
  * one
  * two
  * three

`
  actual := fmt.Sprintf("%s", Flatten(original))

  if expected != actual {
    t.Fatalf("expected: %s, got: %s", expected, actual)
  }
}
```

我们来看下 `Flatten` 是如何实现的，判断当前的 `err` 是不是 `*Error` 类型，如果是的话，就遍历 `err.Errors`，然后递归调用 `flatten` 函数，如果不是 `*Error` 类型的函数，那么就直接将 `err` 添加到 `flatErr.Errors` 中

```go
func Flatten(err error) error {
  flatErr := new(Error)
  flatten(err, flatErr)
  return flatErr
}

func flatten(err error, flatErr *Error) {
  switch err := err.(type) {
  case *Error:
    for _, e := range err.Errors {
      flatten(e, flatErr)
    }
  default:
    flatErr.Errors = append(flatErr.Errors, err)
  }
}
```

我们再来看下一个测试用例 `TestFlatten_nonError`，传入 `Flatten` 的参数是一个原生的 `error` 类型的错误，那么 `Flatten` 函数应该怎么处理呢

```go
func TestFlatten_nonError(t *testing.T) {
  err := errors.New("foo")
  actual := Flatten(err)
  if !reflect.DeepEqual(actual, err) {
    t.Fatalf("bad: %#v", actual)
  }
}
```

在 `Flatten` 中先判断一下 `err` 是不是 `*Error` 类型，如果不是的话，直接把这个错误返回出去

```go
func Flatten(err error) error {
  // If it isn't an *Error, just return the error as-is
  if _, ok := err.(*Error); !ok {
    return err
  }
}
```

## Group

我们在来看第三个接口 `Group`

`Group` 是一个结构体，它有三个属性：

- `mutex`：互斥锁
- `err`：自定义的 `Error` 类型的错误
- `wg`：等待组

```go
type Group struct {
  mutex sync.Mutex
  err   *Error
  wg    sync.WaitGroup
}
```

提供两个方法 `Go` 和 `Wait`

`Go` 函数的作用是在 `goroutine` 中将 `err` 添加到 `err` 中

`Wait` 函数的作用是需要等等到所有的 `goroutine` 都执行完毓，然后返回 `err` 中的信息

测试用例如下：

```go
func TestGroup(t *testing.T) {
  err1 := errors.New("group_test: 1")
  err2 := errors.New("group_test: 2")

  cases := []struct {
    errs      []error
    nilResult bool
  }{
    {errs: []error{}, nilResult: true},
    {errs: []error{nil}, nilResult: true},
    {errs: []error{err1}},
    {errs: []error{err1, nil}},
    {errs: []error{err1, nil, err2}},
  }

  for _, tc := range cases {
    var g Group

    for _, err := range tc.errs {
      err := err
      g.Go(func() error { return err })

    }

    gErr := g.Wait()
    fmt.Println(gErr)
    if gErr != nil {
      for i := range tc.errs {
        if tc.errs[i] != nil && !strings.Contains(gErr.Error(), tc.errs[i].Error()) {
          t.Fatalf("expected error to contain %q, actual: %v", tc.errs[i].Error(), gErr)
        }
      }
    } else if !tc.nilResult {
      t.Fatalf("Group.Wait() should not have returned nil for errs: %v", tc.errs)
    }
  }
}
```

我们来看下 `Go` 和 `Wait` 的实现：

`Go` 函数被调用一次就会调用一次 `wg.Add(1)`，然后在 `goroutine` 中执行 `f` 函数，如果 `f` 函数返回的错误不为空，那么就加锁，将错误添加到 `err` 中，等到 `goroutine` 执行完后，调用 `wg.Done()`

```go
func (g *Group) Go(f func() error) {
  g.wg.Add(1)

  go func() {
    defer g.wg.Done()

    if err := f(); err != nil {
      g.mutex.Lock()
      g.err = Append(g.err, err)
      g.mutex.Unlock()
    }
  }()
}
```

`Wait` 函数被调用时就会调用 `wg.Wait()`，`wg.Wait()` 会等待所有的 `goroutine` 执行完，然后才会执行下面的代码

```go
func (g *Group) Wait() *Error {
  g.wg.Wait()
  g.mutex.Lock()
  defer g.mutex.Unlock()
  return g.err
}
```

具体的源码如下：

```go
type Group struct {
  mutex sync.Mutex
  err   *Error
  wg    sync.WaitGroup
}

func (g *Group) Go(f func() error) {
  g.wg.Add(1)

  go func() {
    defer g.wg.Done()

    if err := f(); err != nil {
      g.mutex.Lock()
      g.err = Append(g.err, err)
      g.mutex.Unlock()
    }
  }()
}

func (g *Group) Wait() *Error {
  g.wg.Wait()
  g.mutex.Lock()
  defer g.mutex.Unlock()
  return g.err
}
```

## Prefix

在来看第三个接口 `Prefix`

`Prefix` 接口的作用是在错误前面加上前缀

```go
func TestPrefix_Error(t *testing.T) {
  original := &Error{
    Errors: []error{errors.New("foo")},
  }

  result := Prefix(original, "bar")
  if result.(*Error).Errors[0].Error() != "bar foo" {
    t.Fatalf("bad: %s", result)
  }
}
```

要实现这个功能也是比较简单的，加前缀的功能使用 `fmt.Errorf` 函数就可以了

```go
func Prefix(err error, prefix string) error {
  newErr := err.(*Error)
  for i, e := range newErr.Errors {
    newErr.Errors[i] = fmt.Errorf("%s %s", prefix, e)
  }
  return newErr
}
```

我们再来看下一个测试用例 `TestPrefix_NilError` 这个测试用例的作用是，如果传入的 `err` 是 `nil` 类型的话，那么返回的结果也是 `nil`

```go
func TestPrefix_NilError(t *testing.T) {
  var err error
  result := Prefix(err, "bar")
  if result != nil {
    t.Fatalf("bad: %#v", result)
  }
}
```

所以在 `Prefix` 函数一进来时，就需要判断下 `err` 是不是为 `nil`，如果为 `nil` 的话，直接返回 `nil` 就可以了

```go
func Prefix(err error, prefix string) error {
  if err == nil {
    return nil
  }
  newErr := err.(*Error)
  for i, e := range newErr.Errors {
    newErr.Errors[i] = fmt.Errorf("%s %s", prefix, e)
  }
  return newErr
}
```

## Sort

`Sort` 接口是由 `sort.Sort 函数提供的，它的作用是进行排序

`sort.Sort` 函数的参数是一个 `sort.Interface` 接口，这个接口有三个方法：

- `Len`：返回切片的长度
- `Less`：比较两个元素的大小
- `Swap`：交换两个元素的位置

所以需要对 `Error` 结构体实现这三个方法

```go
func (e *Error) Len() int {
  if e == nil {
    return 0
  }

  return len(e.Errors)
}

func (e *Error) Swap(i, j int) {
  e.Errors[i], e.Errors[j] = e.Errors[j], e.Errors[i]
}

func (e *Error) Less(i, j int) bool {
  return e.Errors[i].Error() < e.Errors[j].Error()
}
```

这是具体的测试用例：

```go
func TestSortSingle(t *testing.T) {
  errFoo := errors.New("foo")

  expected := []error{
    errFoo,
  }

  err := &Error{
    Errors: []error{
      errFoo,
    },
  }

  sort.Sort(err)
  if !reflect.DeepEqual(err.Errors, expected) {
    t.Fatalf("bad: %#v", err)
  }
}

func TestSortMultiple(t *testing.T) {
  errBar := errors.New("bar")
  errBaz := errors.New("baz")
  errFoo := errors.New("foo")

  expected := []error{
    errBar,
    errBaz,
    errFoo,
  }

  err := &Error{
    Errors: []error{
      errFoo,
      errBar,
      errBaz,
    },
  }

  sort.Sort(err)
  if !reflect.DeepEqual(err.Errors, expected) {
    t.Fatalf("bad: %#v", err)
  }
}
```
