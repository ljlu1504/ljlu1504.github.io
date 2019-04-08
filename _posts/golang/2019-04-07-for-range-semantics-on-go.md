---
layout: post
title: golang:GO的ForRange语义设计
category: golang
tags: golang
description: 本文将探讨Go中for range, 值/指针语义背后的机制和设计
keywords: golang, for range, pointer, value
date: 2019-04-07T15:19:54+08:00
score: 5.0
---

## 原文链接
https://www.ardanlabs.com/blog/2017/06/for-range-semantics.html

## 前言

在阅读这篇文章前会好先阅读这四篇文章，它们将理解这篇文章。

四部分系列索引:
1. GO堆栈和指针
2. GO逃避分析
3. GO内存分析
4. GO数据和语义的设计

在Go编程语言中，值和指针语义的概念无处不在。 如前所述，语义一致性对于完整性和可读性至关重要。 它允许开发人员在代码库不断增长时保持强大的代码库设计思路。 它还有助于最大限度地减少错误，副作用和未知行为。

## 介绍
在这篇文章中，我将探讨Go中的for range语句如何提供值和指针语义形式。 我将教授语言机制，并向您展示这些语义的深层含义。 然后我将展示一个简单的例子，说明混合这些语义和可能导致的问题是多么容易。

## GO语言机制(Language Mechanics)
从这段代码开始，该代码显示for range循环的值语义形式。
[go play](https://play.golang.org/p/_CWCAF6ge3)
```go
01 package main
02
03 import "fmt"
04
05 type user struct {
06     name string
07     email string
08 }
09
10 func main() {
11     users := []user{
12         {"Bill", "bill@email.com"},
13         {"Lisa", "lisa@email.com"},
14         {"Nancy", "nancy@email.com"},
15         {"Paul", "paul@email.com"},
16     }
17
18     for i, u := range users {
19         fmt.Println(i, u)
20     }
21 }
```
在以上示例中，程序声明了一个名为user的类型，创建了四个用户值，然后显示有关每个用户的信息。 第18行的for range循环使用值语义。 这是因为在每次迭代时，来自slice的原始用户值的副本在循环内部进行并操作。 实际上，对Println的调用会创建循环副本的第二个副本。 如果要将值语义用于用户值，那么这就是您想要的。

如果您要使用指针语义，则for range循环将如下所示。
```go
18     for i := range users {
19         fmt.Println(i, users[i])
20     }
```
现在循环已被修改为使用指针语义。 循环内的代码不再在其自己的副本上运行，而是在存储在slice内的原始user 值上运行。 但是，对Println的调用仍然使用值语义并且正在传递副本。
要解决此问题，需要进一步修改。
```go
18     for i := range users {
19         fmt.Println(i, &users[i])
20     }
```
现在，用户数据一直使用指针机制。

作为参考，下面并排显示了值和指针语义。
```go
       // Value semantics.           // Pointer semantics.
18     for i, u := range users {     for i := range users {
19         fmt.Println(i, u)             fmt.Println(i, &users[i])
20     }                             }
```
##Deeper Mechanics
让我们来看看比上面例子更深层次的Go语言机制。请看下面的这个程序。程序初始化一个字符串数组，迭代这些字符串，并在每次迭代时更改索引1处的字符串。
[go play](https://play.golang.org/p/IlAiEkgs4C)
```go
01 package main
02
03 import "fmt"
04
05 func main() {
06     five := [5]string{"Annie", "Betty", "Charley", "Doug", "Edward"}
07     fmt.Printf("Bfr[%s] : ", five[1])
08
09     for i := range five {
10         five[1] = "Jack"
11
12         if i == 1 {
13             fmt.Printf("Aft[%s]\n", five[1])
14         }
15     }
16 }
```
这段程序预计的输出是什么？
```
Bfr[Betty] : Aft[Jack]
```
正如您所料，第10行的代码已更改索引1处的字符串，您可以在输出中看到该字符串。 该程序使用for range循环的`指针`语义版本。 接下来，代码将使用for range循环的`值`语义版本。
[go play](https://play.golang.org/p/opSsIGtNU1)
```go
01 package main
02
03 import "fmt"
04
05 func main() {
06     five := [5]string{"Annie", "Betty", "Charley", "Doug", "Edward"}
07     fmt.Printf("Bfr[%s] : ", five[1])
08
09     for i, v := range five {
10         five[1] = "Jack"
11
12         if i == 1 {
13             fmt.Printf("v[%s]\n", v)
14         }
15     }
16 }
```
在循环的每次迭代中，代码再次更改索引1处的字符串。这次当代码显示索引1处的值时，输出是不同的。
```go
Bfr[Betty] : v[Betty]
```
您可以看到`for range`的这种形式是真正使用值语义。 `for range`是迭代它自己的`数组副本`。 这就是输出中没有看到变化的原因。

当使用值语义形式在slice上进行`ranging`时，`for range`采用`slice header`的副本。 这就是一下代码不会引起panic的原因。
[go play](https://play.golang.org/p/OXhdsneBec)
```go
01 package main
02
03 import "fmt"
04
05 func main() {
06     five := []string{"Annie", "Betty", "Charley", "Doug", "Edward"}
07
08     for _, v := range five {
09         five = five[:2]
10         fmt.Printf("v[%s]\n", v)
11     }
12 }

Output:
v[Annie]
v[Betty]
v[Charley]
v[Doug]
v[Edward]
```
如果你看第`09`行，切片值在循环内减少到2的长度，但是循环在它自己的`slice`副本上运行。 这允许循环使用原始长度进行迭代而没有任何问题，因为`底层Array`并没有被修改。


如果代码使用`for range`的指针语义形式，则程序会发生`panics`。
[go play](https://play.golang.org/p/k5a73PHaka)
```go
01 package main
02
03 import "fmt"
04
05 func main() {
06     five := []string{"Annie", "Betty", "Charley", "Doug", "Edward"}
07
08     for i := range five {
09         five = five[:2]
10         fmt.Printf("v[%s]\n", five[i])
11     }
12 }

Output:
v[Annie]
v[Betty]
panic: runtime error: index out of range

goroutine 1 [running]:
main.main()
    /tmp/sandbox688667612/main.go:10 +0x140
```
`for range`在迭代之前使用slice的长度，但在循环期间，slice长度发生了变化。 在第三次迭代中，循环尝试访问不再与slice长度相关联的元素。

##Mixing Semantics
这是一个反例。 此代码混合了user类型的语义，并导致错误。

[go play](https://play.golang.org/p/L_WmUkDYFJ)
```go
01 package main
02
03 import "fmt"
04
05 type user struct {
06     name  string
07     likes int
08 }
09
10 func (u *user) notify() {
11     fmt.Printf("%s has %d likes\n", u.name, u.likes)
12 }
13
14 func (u *user) addLike() {
15     u.likes++
16 }
17
18 func main() {
19     users := []user{
20         {name: "bill"},
21         {name: "lisa"},
22     }
23
24     for _, u := range users {
25         u.addLike()
26     }
27
28     for _, u := range users {
29         u.notify()
30     }
31 }
```
这个例子并不是专门设计的。 在第`05`行，声明`user`类型，并选择指针语义来实现`user`类型的方法集。 然后在`main`程序中，在`for range`循环中使用值语义来向每个用户添加`like`。 然后使用第二个循环来再次使用值语义来通知每个`user`。
```go
bill has 0 likes
lisa has 0 likes
```
输出显示没有添加任何`like`。 我不能强调你应该为给定类型选择一个语义，并坚持使用该类型的数据。

以下代码展示了user类型的指针语义保持一致的用法。
[go play](https://play.golang.org/p/GwAnyBNqPz)
```go
01 package main
02
03 import "fmt"
04
05 type user struct {
06     name  string
07     likes int
08 }
09
10 func (u *user) notify() {
11     fmt.Printf("%s has %d likes\n", u.name, u.likes)
12 }
13
14 func (u *user) addLike() {
15     u.likes++
16 }
17
18 func main() {
19     users := []user{
20         {name: "bill"},
21         {name: "lisa"},
22     }
23
24     for i := range users {
25         users[i].addLike()
26     }
27
28     for i := range users {
29         users[i].notify()
30     }
31 }

// Output:
bill has 1 likes
lisa has 1 likes
```

## 结论
值和指针语义是`Go`编程语言的重要组成部分，正如我所示，它集成到`for range`循环中。 使用`for range`时，确保您正在对给定迭代的类型使用正确的格式。最后想要说的是如果你没有加倍注意，用`for range`很容易导致语义的混用。

`Go`语言赋予您选择语义的能力，并且可以干净利落地使用它。 这是你想要充分利用的东西。 我希望您确定每种类型使用的语义并保持一致。 您对一段数据的语义越一致，您的代码库就越好。 如果您有充分的理由更改语义，请进行广泛的记录。