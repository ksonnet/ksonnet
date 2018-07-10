#!/usr/bin/env ruby

# verify charts specified in a helm repository render with ks
# usage: ks-verify-helm.rb [helm repo URL]

require 'net/http'
require 'open3'
require 'timeout'
require 'tmpdir'
require 'uri'
require 'yaml'

DEFAULT_HELM_REPO_URL="https://kubernetes-charts.storage.googleapis.com/index.yaml"

helm_repo_url = ARGV[0] || DEFAULT_HELM_REPO_URL

tmp_dir = Dir.mktmpdir("ks-verify-helm-", "/tmp")
at_exit { FileUtils.remove_entry(tmp_dir) }

app_name = (0...6).map { (65 + rand(26)).chr }.join

# generate a ksonnet app
Dir.chdir(tmp_dir) do
    `ks init #{app_name} --skip-default-registries > /dev/null 2>&1`
end

fails = []

Dir.chdir("#{tmp_dir}/#{app_name}") do
    uri = URI.parse(helm_repo_url)
    response = Net::HTTP.get(uri)

    `ks registry add helm #{helm_repo_url}`

    repo = YAML.load(response)
    repo["entries"].each_with_index do |(key,_), index|
        puts "checking #{index+1}: #{key}"
        `ks pkg install helm/#{key} > /dev/null 2>&1`
        `ks module create #{key} > /dev/null 2>&1`
        `ks env add #{key} > /dev/null 2>&1`
        `ks env targets #{key} --module #{key} > /dev/null 2>&1`

        Open3.popen3("ks generate helm-#{key} #{key}.#{key}") do |_, _, stderr, thr|
            if thr.value != 0
                puts "generating #{key} failure:"
                puts stderr.read
                fails << key
                next
            end
        end

        begin
            Timeout::timeout(15) do
                Open3.popen3("ks show #{key}") do |_, _, stderr, thr|
                    if thr.value != 0
                        puts "#{key} failed"
                        puts stderr.read
                        fails << key
                    end
                end
            end
        rescue => e
            puts "timed out showing #{key}"
        end
    end
end

unless fails.empty?
    puts fails.inspect
    exit(1)
end


