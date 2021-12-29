## Run the crawl from the Docker image
You can run a crawl from the packaged Docker image to crawl your website. You will need to install jq, a lightweight command-line JSON processor

Then you need to start the crawl according to your configuration. You should check the dedicated configuration documentation.

docker run -it --env-file=.env -e "CONFIG=$(cat config.json | jq -r tostring)" algolia/docsearch-scraper
