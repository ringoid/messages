stage-all: clean stage-deploy
test-all: clean test-deploy
prod-all: clean prod-deploy

build:
	@echo '--- Building internal-get-messages-messages function ---'
	GOOS=linux go build lambda-internal-getmessages/get_messages.go
	@echo '--- Building warmup-messages function ---'
	GOOS=linux go build lambda-warmup/warm_up.go
	@echo '--- Building internal-handle-stream-messages function ---'
	GOOS=linux go build lambda-handle-stream/handle_stream.go lambda-handle-stream/message.go

zip_lambda: build
	@echo '--- Zip internal-get-messages-messages function ---'
	zip get_messages.zip ./get_messages
	@echo '--- Zip warmup-messages function ---'
	zip warmup.zip ./warm_up
	@echo '--- Zip internal-handle-stream-messages function ---'
	zip handle_stream.zip ./handle_stream


test-deploy: zip_lambda
	@echo '--- Build lambda test ---'
	@echo 'Package template'
	sam package --template-file messages-template.yaml --s3-bucket ringoid-cloudformation-template --output-template-file messages-template-packaged.yaml
	@echo 'Deploy test-messages-stack'
	sam deploy --template-file messages-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name test-messages-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=test --no-fail-on-empty-changeset

stage-deploy: zip_lambda
	@echo '--- Build lambda stage ---'
	@echo 'Package template'
	sam package --template-file messages-template.yaml --s3-bucket ringoid-cloudformation-template --output-template-file messages-template-packaged.yaml
	@echo 'Deploy stage-feeds-stack'
	sam deploy --template-file messages-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name stage-messages-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=stage --no-fail-on-empty-changeset

prod-deploy: zip_lambda
	@echo '--- Build lambda prod ---'
	@echo 'Package template'
	sam package --template-file messages-template.yaml --s3-bucket ringoid-cloudformation-template --output-template-file messages-template-packaged.yaml
	@echo 'Deploy prod-feeds-stack'
	sam deploy --template-file messages-template-packaged.yaml --s3-bucket ringoid-cloudformation-template --stack-name prod-messages-stack --capabilities CAPABILITY_IAM --parameter-overrides Env=prod --no-fail-on-empty-changeset

clean:
	@echo '--- Delete old artifacts ---'
	rm -rf get_messages
	rm -rf get_messages.zip
	rm -rf warmup-image.zip
	rm -rf warm_up
	rm -rf handle_stream.zip
	rm -rf handle_stream

