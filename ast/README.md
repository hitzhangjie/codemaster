# AST

AST，是抽象语法树（Abstract Semantic Tree）的简称，符合语法Specification的源代码都可以成功构建出一棵AST。
借助AST我们可以做一些“好玩”的事情，以下是之前整理的一点笔记，cp过来让单薄的测试代码显得丰满一点 :)

## 2020.10.15 AST抽象语法树

前些天阅读go程序源代码时，注意到`go fix`的实现是基于go/ast来实现的。此前不久看过一个babygo项目也是利用go/ast来实现了简单的语法分析，进而实现完整的编译过程。我对此产生了一点兴趣，想进一步研究下go/ast的实现以及用途。

先贴一个我写的文章，是关于微服务代码逻辑可视化的：<https://hitzhangjie.github.io/blog/2020-10-06-visualizing-your-go-code，这里其实也使用了go/ast的能力。文章中也列举了一些业界的用法。>

- 对代码进行静态检查，如遗漏的error处理，goroutine未捕获可能的panic；
- 对代码风格进行检查，检查导出类型、函数、方法有没有添加godoc注释；
- 对代码中的类型进行分析，自动化构建uml类图，如goplantuml；
- 对代码中的控制逻辑、rpc调用进行分析，构建时序图，如ballerina框架；
- 对微服务体系中的rpc调用关系进行分析，构建时序图，如devapi.test；
- 等等；

我们对源程序执行编译构建动作，主要包括如下步骤：

- 词法分析，将源程序中的文字解析为token序列；
- 语法分析，将token序列按照语言文法进行分析，检查是否符合语言的文法，如token序列可能形成了表达式、语句；
- 语义分析，对结构上正确的源程序进行更多的上下文相关的检查，如检查赋值语句`a int = "hello"`中数据类型是否匹配，并报告错误。
       语法分析，从更宽泛的概念上讲，包含了语义分析的过程

- 中间代码生成
- 代码优化
- 目标代码生成

## 2020.10.24 AST VS. CST

语法树类型包括CST、AST，CST可以理解成更加具体的语法树，AST则是对CST进行简化后得到的，AST更简单适合用来对代码进行分析。
在CST里面，每个node都有单独的类型表示，但是在AST里面不是，只有一种类型syntax.Node，通过一个字段来区分不同的类型，更简单。
关于二者的想起区别，可以参考这篇文章，[Abstract VS. Concrete Syntax Tree](https://eli.thegreenplace.net/2009/02/16/abstract-vs-concrete-syntax-trees)。

## 工欲善其事：AST可视化工具

想更加快速深入地了解Go AST，一个好用的AST分析工具是少不了的，我之前收藏过两个：

- [GoAST Viewer](https://yuroyoro.github.io/goast-viewer/index.html)
- [AST Explorer](https://astexplorer.net/)

你可以直接在输入框里面输入go源码（可以是不完整代码），然后看对应的AST是长什么样子的。
并且，AST Explorer还允许我们在源码中选中一个或者部分代码，然后高亮显示其对应的AST部分是长什么样的，非常方便。