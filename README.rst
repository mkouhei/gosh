====================================
 Gosh: interactive shell for golang
====================================

``Gosh`` is the interactive Golang shell.
The goal is to provide an easy-to-use interactive execution environment.

.. image:: https://secure.travis-ci.org/mkouhei/gosh.png
   :target: http://travis-ci.org/mkouhei/gosh
.. image:: https://coveralls.io/repos/mkouhei/gosh/badge.png?branch=master
   :target: https://coveralls.io/r/mkouhei/gosh?branch=master

Features
--------

* Interactive shell
* Enable to omit the main function
* Enable to omit package statement
* Enable to omit the import statement of standard library
* Enable to Import libraries of non-standard library
* Enable to re-declare function
* Ignoring duplicate import package
* Ignoring unused import package

Requirements
------------

* Golang >= 1.2
* `goimports <http://godoc.org/code.google.com/p/go.tools/cmd/goimports>`_ command

Installation
------------

Debian
~~~~~~

Install the follow packages

* golang
* golang-go.tools


In the case using Golang not-distribution package,
execute next command.::

  $ go get code.google.com/p/go.tools/cmd/goimports

Set ``GOPATH``, and execute follows.::

  $ go get github.com/mkouhei/gosh
  
Basic usage
-----------

Examples::

  $ $GOPATH/bin/gosh
  >>> import "fmt"
  >>> func main() {
  >>> fmt.Println("hello")
  >>> }
  hello
  >>>

or::

  $ $GOPATH/bin/gosh
  >>> func main() {
  >>> fmt.Println("hello")
  >>> }
  hello
  >>>

Enable to omit import statement related with standard libraries.

Enable to Import libraries of non-standard library
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

For example of using the some package.::

  >>> import "net/http"
  >>> import "example.org/somepkg"
  >>> func main() {
  >>> r, _ := http.Get("http://example.org/some")
  >>> defer r.Body.Close()
  >>> p, _ := somepkg.Reader(r.Body)
  >>> fmt.Println(p)
  >>> }
  
  (print some payload)

Usage when omitting main function declarations
----------------------------------------------

Example::

  $ $GOPATH/bin/gosh
  >>> i := 1
  >>> i++
  >>> fmt.Println(i)
  2
  >>>

Terminate gosh to reset main declarations, or declare func main without body.::

  $ $GOSH/bin/gosh
  >>> i := i
  >>> fmt.Println(i)
  1
  >>> func main() {}
  >>> fmt.Println(i)
  [error] # command-line-arguments
  ./gosh_tmp.go:8: undefined: i
  >>>

Limitations
~~~~~~~~~~~

* ``fmt.Print*`` are executed only once.

Known issues
~~~~~~~~~~~~

Not evaluate when there are declared and not used valiables.::

  $ $GOPATH/bin/gosh
  >>> i := 1
  >>> fmt.Println("hello")
  >>>


Roadmap
-------

* Tab completion
* Enable to omit import statement of system global installed packages

License
-------

``Gosh`` is licensed under GPLv3.
