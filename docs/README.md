# How to write doc in local

## Prepare
1. Install Python 3.X with zlib, libssl-dev(openssl-devel)
1. Install pip3
1. Install sphinx(>=2.0.0) support https://docs.readthedocs.io/en/latest/getting_started.html
1. Install RTD module & recommonmark
```shell
sudo pip install sphinx_rtd_theme
sudo pip install recommonmark
```

## Generate doc

In windows
```shell
cd docs
make.bat html
```

In linux
```shell
cd docs
make html 
```

## Check the result

1. See html pages in _build folder
2. You can start a http server using `python -m http.server` which will serve at http://0.0.0.0:8000/.
