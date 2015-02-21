====================================
 Gosh: interactive shell for golang
====================================

``Gosh`` is the interactive Golang shell.
The goal is to provide an easy-to-use interactive execution environment.

.. image:: https://secure.travis-ci.org/mkouhei/gosh.png
   :target: http://travis-ci.org/mkouhei/gosh
.. image:: https://coveralls.io/repos/mkouhei/gosh/badge.png?branch=master
   :target: https://coveralls.io/r/mkouhei/gosh?branch=master
.. image:: https://readthedocs.org/projects/gosh/badge/?version=latest
   :target: https://readthedocs.org/projects/gosh/?badge=latest
   :alt: Documentation Status

Documentation
=============

http://gosh.readthedocs.org/

Features
--------

* Interactive shell
* Enable to omit the main function
* Enable to omit package statement
* Enable to omit the import statement of standard library
* Enable to Import libraries of non-standard library
* Enable to re-declare function, type
* Ignoring duplicate import package
* Ignoring unused import package

Requirements
------------

* Golang >= 1.2
* `goimports <http://godoc.org/code.google.com/p/go.tools/cmd/goimports>`_ command

  * We recommend that you install ``goimports`` to ``$PATH`` in advance.
  * Installing automatically if the command is not found in ``$PATH`` (>= v0.3.0).
  * However, the time until the installation is complete in this case,
    you will be waiting for the launch of "``Gosh``" process.

Installation
------------

Debian
~~~~~~

Install the follow packages

* golang
* golang-go.tools (recommended)

Set ``GOPATH``::

  $ install -d /path/to/gopath
  $ export GOPATH=/path/to/gopath

If you install ``goimports`` in advance (recommended),::

  $ sudo apt-get install -y golang-go.tools

Install ``Gosh`` to ``GOPATH``.::

  $ go get github.com/mkouhei/gosh


OS X
~~~~

Install the follow packages with `Homebrew <http://brew.sh/>`_.

* Go
* Mercurial (with Homebrew)

Set ``GOPATH``,::

  $ install -d /path/to/gopath
  $ export GOPATH=/path/to/gopath

If you install ``goimports`` in advance (recommend),::

  $ export PATH=${GOPATH}/bin:$PATH
  $ go get code.google.com/p/go.tools/cmd/goimport

Install the ``Gosh``,::

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
