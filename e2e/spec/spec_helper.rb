require 'bundler/setup'
require 'awspec'
require 'aws-sdk-core'
require 'aws-sdk-sts'
require 'aws-sdk-apigatewayv2'
require 'aws-sdk-ssm'
require 'git'
require 'dotenv'

ENV["AWS_REGION"] = "us-west-2"

RSpec.configure do |config|
  config.expect_with :rspec do |expectations|
    expectations.include_chain_clauses_in_custom_matcher_descriptions = true
  end

  config.mock_with :rspec do |mocks|
    mocks.verify_partial_doubles = true
  end

  config.shared_context_metadata_behavior = :apply_to_host_groups

  # Set formatter regardless of number of files
  config.default_formatter = "doc"
end

RSpec.shared_examples 'common tag tests' do |expected|
  describe "tags" do
    it { should have_tag('Branch').value(expected.branch) }
    it { should have_tag('Sha').value(expected.sha) }
    it { should have_tag('Origin').value(expected.origin) }
    it { should have_tag('Dirty').value(expected.dirty.to_s) }
  end
end

RSpec.shared_examples 'common lambda env tests' do |expected|
  describe "env" do
    {
      "GIT_BRANCH" => expected.branch,
      "GIT_SHA" => expected.sha,
      "GIT_ORIGIN" => expected.origin,
    }.each do |key, value|
      context "#{key}" do
        it { should have_env_var_value(key, value) }
      end
    end
  end
end

module Monad
  def self.substrate
    raise "SUBSTRATE_NAME is required" if ENV["SUBSTRATE_NAME"].nil?
    raise "SUBSTRATE_FEATURES is required" if ENV["SUBSTRATE_FEATURES"].nil?

    @substrate ||= Substrate.new(ENV["SUBSTRATE_NAME"], ENV["SUBSTRATE_FEATURES"])
  end

  def self.expected
    raise "FUNCTION_PATH is required" if ENV["FUNCTION_PATH"].nil?

    @expected ||= Expect.new(ENV["FUNCTION_PATH"], self.substrate)
  end
end


class Substrate
  def initialize(substrate, features)
    raise "substrate is required" if substrate.nil?
    raise "features is required" if features.nil?

    @ssm_client = Aws::SSM::Client.new
    @substrate = substrate
    @features = features.split(',')
    @env = { **self.substrate, **self.features }
  end

  def fetch(path)
    Dotenv::Parser.call(@ssm_client.get_parameter({
      name: path,
      with_decryption: true
    }).parameter.value)
  end

  def substrate
    fetch("/bkeane/substrate/#{@substrate}")
  end

  def features
    @features.reduce({}) do |acc, feature|
      acc.merge(fetch("/bkeane/substrate/#{@substrate}/#{feature}"))
    end
  end

  def gateway_name
    if self.apigateway_enabled?
      client = Aws::ApiGatewayV2::Client.new
      client.get_api(api_id: @env["SUBSTRATE_APIGATEWAY_ID"]).name
    end
  end

  def gateway_endpoint
    if self.apigateway_enabled?
      client = Aws::ApiGatewayV2::Client.new
      api_id = @env["SUBSTRATE_APIGATEWAY_ID"]
      
      # Get all domain names and find the one mapped to our API
      domain_names = client.get_domain_names.items
      domain_name = domain_names.find do |domain|
        mappings = client.get_api_mappings({domain_name: domain.domain_name}).items
        mappings.any? { |mapping| mapping.api_id == api_id }
      end

      # Return the domain name URL if found, otherwise fall back to default endpoint
      domain_name ? "https://#{domain_name.domain_name}" : client.get_api(api_id: api_id).api_endpoint
    end
  end

  def eventbridge_enabled?
    @env.key?("SUBSTRATE_EVENTBRIDGE_ENABLE") && @env["SUBSTRATE_EVENTBRIDGE_ENABLE"] == "true"
  end

  def apigateway_enabled?
    @env.key?("SUBSTRATE_APIGATEWAY_ENABLE") && @env["SUBSTRATE_APIGATEWAY_ENABLE"] == "true"
  end

  def vpc_enabled?
    @env.key?("SUBSTRATE_VPC_SUBNET_IDS") && @env.key?("SUBSTRATE_VPC_SECURITY_GROUP_IDS")
  end

  def no_auth_enabled?
    !@env.key?("SUBSTRATE_APIGATEWAY_AUTH_TYPE")
  end

  def aws_auth_enabled?
    @env.key?("SUBSTRATE_APIGATEWAY_AUTH_TYPE") && @env["SUBSTRATE_APIGATEWAY_AUTH_TYPE"] == "AWS_IAM"
  end

  def jwt_auth_enabled?
    @env.key?("SUBSTRATE_APIGATEWAY_AUTH_TYPE") && @env["SUBSTRATE_APIGATEWAY_AUTH_TYPE"] == "JWT"
  end

  def option_prefix_paths_with_org?
    @env.key?("SUBSTRATE_PREFIX_PATHS_WITH_ORG") && @env["SUBSTRATE_PREFIX_PATHS_WITH_ORG"] == "true"
  end

  def option_prefix_names_with_org?
    @env.key?("SUBSTRATE_PREFIX_NAMES_WITH_ORG") && @env["SUBSTRATE_PREFIX_NAMES_WITH_ORG"] == "true"
  end

  def env
    @env
  end
end

class Expect
  attr_reader :path, :git

  def initialize(function_path, substrate)
    raise "function_path is required" if function_path.nil?
    raise "substrate is required" if substrate.nil?

    @caller = Aws::STS::Client.new.get_caller_identity
    @substrate = substrate
    @path = function_path
    @git = Git.open(function_path)
  end

  def env
    @substrate.env
  end

  def region
      "us-west-2"
  end

  def name
      File.basename(@path)
  end

  def org
      @git.remote.url.match(/github\.com[:\/](.+?)\/(.+?)\.git/)[1]
  end

  def repo
      @git.remote.url.match(/github\.com[:\/](.+?)\/(.+?)\.git/)[2]
  end

  def origin
      "github.com/#{self.org}/#{self.repo}"
  end

  def branch
      @git.current_branch
  end

  def sha
      @git.revparse("HEAD")
  end

  def dirty
    !@git.status.changed.empty?
  end

  def resource_name
      if @substrate.option_prefix_names_with_org?
          "#{self.org}-#{self.repo}-#{self.branch}-#{self.name}"
      else
          "#{self.repo}-#{self.branch}-#{self.name}"
      end
  end

  def resource_path
      if @substrate.option_prefix_paths_with_org?
          "#{self.org}/#{self.repo}/#{self.branch}/#{self.name}"
      else
          "#{self.repo}/#{self.branch}/#{self.name}"
      end
  end

  def function_name
      self.resource_name
  end

  def function_arn
      "arn:aws:lambda:#{self.region}:#{@caller.account}:function:#{self.function_name}"
  end

  def role_name
      self.resource_name
  end

  def role_arn
      "arn:aws:iam::#{@caller.account}:role/#{self.role_name}"
  end

  def policy_name
      self.resource_name
  end

  def policy_arn
      "arn:aws:iam::#{@caller.account}:policy:#{self.policy_name}"
  end

  def bus_name
      self.env["SUBSTRATE_EVENTBRIDGE_BUS_NAME"]
  end

  def vpc_subnet_ids
      if @substrate.vpc_feature_enabled
          self.env["SUBSTRATE_VPC_SUBNET_IDS"].split(",")
      else
          []
      end
  end

  def vpc_security_group_ids
      if @substrate.vpc_feature_enabled
          self.env["SUBSTRATE_VPC_SECURITY_GROUP_IDS"].split(",")
      else
          []
      end
  end

  def url
      "#{@substrate.gateway_endpoint}/#{self.resource_path}/"
  end

  def route_key
      "ANY /#{self.resource_path}/{proxy+}"
  end
end

def buses(function_path)
  paths = Dir.glob("#{function_path}/bus/*")
  paths.each do |file_path|
      rule = File.basename(file_path).split(".")[0]
      yield(rule) if block_given?
  end
end

def sigv4_get_request(url)
  uri = URI(url)

  signer = Aws::Sigv4::Signer.new(
    service: 'execute-api',
    region: 'us-west-2',
    credentials_provider: Aws::SharedCredentials.new
  )

  request = Net::HTTP::Get.new(uri)
  signed_request = signer.sign_request(http_method: 'GET', url: uri.to_s)

  signed_request.headers.each { |key, value| request[key] = value }

  Net::HTTP.start(uri.host, uri.port, use_ssl: true) do |http|
    http.request(request)
  end
end

def bearer_get_request(url)
  ssmc = Aws::SSM::Client.new
  creds = ssmc.get_parameter({
    name: "/bkeane/test/auth0/client_credentials",
    with_decryption: true
  })


  auth0_uri = URI("https://kaixo.us.auth0.com/oauth/token")
  http = Net::HTTP.new(auth0_uri.host, auth0_uri.port)
  http.use_ssl = true
  http.verify_mode = OpenSSL::SSL::VERIFY_NONE

  request = Net::HTTP::Post.new(auth0_uri)
  request["content-type"] = 'application/json'
  request.body = creds.parameter.value

  response = http.request(request)
  token = JSON.parse(response.read_body)["access_token"]

  uri = URI(url)
  request = Net::HTTP::Get.new(uri)
  request["Authorization"] = "Bearer #{token}"
  Net::HTTP.start(uri.host, uri.port, use_ssl: true) do |http|
    http.request(request)
  end
end