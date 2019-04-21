# Building the docs

```bash
pip install virtualenvwrapper
source /usr/local/bin/virtualenvwrapper.sh
(mkvirtualenv -r requirements.txt docs)
workon docs
make html
google-chrome _build/html/index.html
````

To have Travis deploy to Github Pages, the GITHUB_USER and GITHUB_TOKEN need to be set in either the Travis dashboard
or in a Travis encrypted environment variable eg: 

```bash
travis encrypt GH_TOKEN=github_user:1230000000000000000000000000000000000000 --com --add env.matrix
```
