=============
 User Manual
=============

Run Gosh
========

Check version
---------------
::

   $GOPATH/bin/gosh -version
   version: v0.x.x

Basic mode
----------
::

   $GOPATH/bin/gosh
   go version go1.3.3 linux/amd64
   
   Gosh v0.x.x
   Copyright (C) 2014,2015 Kouhei Maeda
   License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>.
   This is free software, and you are welcome to redistribute it.
   There is NO WARRANTY, to the extent permitted by law.
   >>> 

Debug mode
----------
::

   $GOPATH/bin/gosh -d

Terminate Gosh
==============

Enter Ctrl+D::

  >>> [gosh] terminated
  $

Execute main function
=====================

Go syntax validly::

  >>> package main
  >>> import "fmt"
  >>> func main() {
  >>> fmt.Println("hello")
  >>> }
  hello
  >>>

Omit ``package``, ``import`` statement, ``main func``
-----------------------------------------------------

Gosh supports omitting as follows;

* "``package``" statement
* "``import``" statement for standard libraries
* "``func main``" signature

So users give the same results with the following.::

  >>> fmt.Println("hello")
  hello
  >>>

``fmt.Print*`` are executed only once
-------------------------------------
::

   >>> i:=1
   >>> for i < 3 {
   >>> fmt.Println(i)
   >>> i++
   >>> }

This omit ``main func`` is equivalent to the main func not following omitted.::

  >>> func main() {
  >>> i:=1
  >>> for i < 3 {
  >>> fmt.Println(i)
  >>> i++
  >>> }

But, ``fmt.Print*`` are executed only once.::

  >>> fmt.Println(1)
  1
  >>> fmt.Println(2)
  2

This ``fmt.Print*`` are removed main body after executing main function.


Reset declaration of main func
==============================

Execute follow command.::

  >>> func main() {}

For example, test function(),::

  >>> func test() {
  >>> fmt.Println("hello")
  >>> }

Execute test() twice,::

  >>> test()
  hello
  >>> test()
  hello
  hello

This is equivalent to the main func not following omitted.::

  >>> func main() {
  >>> test()
  >>> test()
  >>> }

So, print "hello" once after reset main.::

  >>> test()
  hello
  >>> func main() {}
  >>> test()
  hello

Import packages
===============

Gosh supports imports 3rd party libraryies. Gosh enter the ``import "package"``, Gosh executes ``go get`` and installs the package into the ``$GOPATH`` of Gosh process.

For example of using the some package.::

  >>> import "example.org/somepkg"
  >>> resp, _ := http.Get("http://example.org/some")
  >>> defer resp.Body.Close()
  >>> payload, _ := somepkg.Reader(resp.Body)
  >>> fmt.Println(payload)
  (print some payload)

Users are able to omit import "``net/http``" package that is Go standard library.

If users import the same package, Gosh ignores duplicate import, adn treats as import of only once.

Declaration of type
===================

Gosh supoorts declaration of type.::

  >>> type foo struct {
  >>> msg string
  >>> cnt int
  >>> }
  >>> f := foo{"hello", 0}
  >>> for f.cnt < 3 {
  >>> fmt.Println(f.msg)
  >>> f.cnt++
  >>> }
  hello
  hello
  hello
  >>>

Gosh supports re-declarations of type. (>= v0.2.3-6-g415df66)

Declaration of function
=======================

Gosh supports declaration of function.::

  >>> func test(msg string) bool {
  >>> if strings.HasPrefix(msg, "Hello") {
  >>> return true
  >>> }
  >>> return false
  >>> }
  >>> fmt.Println(test("helo"))
  false
  >>> fmt.Println(test("hello"))
  false
  >>> fmt.Println(test("Hello"))
  true

Gosh supports re-declarations of function.::

  >>> func bar() {
  >>> fmt.Println("hello")
  >>> }
  >>> bar()
  hello
  >>> func bar() {
  >>> fmt.Println("bye")
  >>> }
  >>> bar()
  bye
  bye
