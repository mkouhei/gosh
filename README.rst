====================================
 Gosh: interactive shell for golang
====================================

``Gosh`` is the interactive Golang shell.
The goal is to provide an easy-to-use interactive execution environment.

.. image:: https://secure.travis-ci.org/mkouhei/gosh.png
   :target: http://travis-ci.org/mkouhei/gosh

Features
--------

* Interactive shell
* Enable to omit package statement
* Enable to omit the import statement of standard library
* Enable to Import libraries of non-standard library
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

Install the follow packages::

* golang
* golang-go.tools


In the case using Golang not-distribution package,
execute next command.::

  $ go get code.google.com/p/go.tools/cmd/goimports

Set ``GOPATH``, and execute follows.::

  $ go get github.com/mkouhei/gosh
  
Usage
-----

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

Roadmap
-------

* Omitting the main function
* Tab completion
* Enable to omit import statement of system global installed packages

License
-------

``Gosh`` is licensed under GPLv3.
