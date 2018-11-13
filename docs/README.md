# Building the docs

```bash
pip install virtualenvwrapper
source /usr/local/bin/virtualenvwrapper.sh
(mkvirtualenv -r requirements.txt docs)
workon docs
make html
google-chrome _build/html/index.html
````
