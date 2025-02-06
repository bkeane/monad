#!/usr/bin/env ruby

require 'optparse'
require 'pry'
class TestRunner
  DEFAULT_OPTIONS = {
    profile: 'personal',
    substrate: 'ager',
    features: '',
    container: true
  }.freeze

  def initialize
    @options = DEFAULT_OPTIONS.dup
  end

  def run
    parse_options
    validate_arguments
    setup_environment
    execute_tests
  end

  private

  def parse_options
    OptionParser.new do |opts|
      opts.banner = "Usage: #{$0} [options] function_path"

      opts.on('-p', '--profile PROFILE', 'AWS profile to use [personal]') do |profile|
        @options[:profile] = profile
      end

      opts.on('-n', '--substrate SUBSTRATE', 'Substrate to test [platform]') do |substrate|
        @options[:substrate] = substrate
      end

      opts.on('-f', '--features FEATURES', Array, 'Comma-separated list of features to test (e.g. api,auth0) []') do |features|
        @options[:features] = features
      end

      opts.on('-c', '--container [BOOL]', TrueClass, 'Run tests in container [true]') do |container|
        @options[:container] = container.nil? ? true : container
      end

      opts.on('-h', '--help', 'Show this help message') do
        puts opts
        exit
      end
    end.parse!

    @options[:function_path] = ARGV[0]
  end

  def validate_arguments
    unless @options[:function_path]
      puts "Error: function_path is required"
      exit 1
    end
  end

  def setup_environment
    ENV['AWS_PROFILE'] = @options[:profile]
    ENV['FUNCTION_PATH'] = @options[:function_path]
    ENV['SUBSTRATE_NAME'] = @options[:substrate]
    ENV['SUBSTRATE_FEATURES'] = Array(@options[:features]).join(',')
  end

  def execute_tests
    if @options[:container]
      system('docker compose run --rm --remove-orphans test')
    else
      system('bundle exec rspec')
    end
  end
end

TestRunner.new.run


