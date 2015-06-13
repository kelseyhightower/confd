require "confd/template/generator/version"
require 'json'

module Confd
  class TemplateGenerator
    attr_reader :input
    def initialize(input)
      @input = input
    end

    def to_tmpl
      recur_to_tmpl(input, {}, [])
    end

    def to_toml
      recur_to_toml(input, [], [])
    end

    def to_json
      recur_to_json(input, {}, []).to_json
    end

    private
    def recur_to_json(input, output, current_keys)
      input.each do |k, v|
        if v.is_a?(Hash)
          recur_to_json(v, output, current_keys + [k])
        else
          keys = current_keys + [k]
          output.merge!({"#{keys.join(".")}" => v})
        end
      end
      output
    end

    def recur_to_tmpl(input, output, current_keys)
      input.each do |k, v|
        output[k] ||= {}
        if v.is_a?(Hash)
          output[k].merge!(recur_to_tmpl(v, output[k], current_keys + [k]))
        else
          keys = current_keys + [k]
          output.merge!({k => %Q[{{getv "/#{keys.join(".")}"}}]})
        end
      end
      output
    end

    def recur_to_toml(input, current_keys, total_keys)
      input.each do |k, v|
        if v.is_a?(Hash)
          recur_to_toml(v, current_keys + [k], total_keys)
        else
          keys = current_keys + [k]
          total_keys << "/#{keys.join(".")}" 
        end
      end
      total_keys
    end
  end
end
