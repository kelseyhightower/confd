require 'json'

module Confd
  class TemplateGenerator
    REGEX = / /
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
      hash = recur_to_json(input, {}, [])
      nil2empty(hash)
      hash.to_json
    end

    private
    def nil2empty(hash)
      hash.keys.each do |key|
        if hash[key].kind_of? Hash
          nil2empty(hash[key])
        else
          hash[key] = '' if hash[key].nil?
        end
      end
    end

    def terminal_arry?(arry)
      arry.all? { |value| [String, Fixnum, Symbol].include?(value.class) }
    end
  
    def recur_to_json(input, output, current_keys)
      if input.is_a?(Array)
        input.each_with_index do |v, k|
          if v.is_a?(Hash) || v.is_a?(Array)
            recur_to_json(v, output, current_keys + [k])
          else
            keys = current_keys + [k]
            output.merge!({"#{keys.join(".")}" => v})
          end
        end
      else
        input.each do |k, v|
          if v.is_a?(Array) && terminal_arry?(v)
            keys = current_keys + [k]
            output.merge!({"#{keys.join(".")}" => v})
          elsif v.is_a?(Hash) || v.is_a?(Array)
            recur_to_json(v, output, current_keys + [k])
          else
            keys = current_keys + [k]
            output.merge!({"#{keys.join(".")}" => v})
          end
        end
      end
      output
    end

    def recur_to_tmpl(input, output, current_keys)
      if input.is_a?(Array)
        arry = []
        input.each_with_index do |v, k|
          if v.is_a?(Hash)  
            arry << (recur_to_tmpl(v, {}, current_keys+[k]))
          else
            keys = current_keys + [k]
            if v =~ REGEX
              arry << %Q["{{getv "/#{keys.join(".")}"}}"]
            else
              arry << %Q[{{getv "/#{keys.join(".")}"}}]
            end
          end
        end
        return arry
      else
        input.each do |k, v|
          output[k] ||= {}
          if v.is_a?(Hash)
            output[k].merge!(recur_to_tmpl(v, output[k], current_keys + [k]))
          elsif v.is_a?(Array) && !terminal_arry?(v)
            output[k] = (recur_to_tmpl(v, output[k], current_keys + [k]))
          elsif v.is_a?(Array) && terminal_arry?(v)
            keys = current_keys + [k]
            output[k] = %Q[{{range jsonArray (getv "/#{keys.join(".")}")}}\n - {{.}}\n {{end}}\n]
          else
            keys = current_keys + [k]
            if v =~ REGEX
              output.merge!({k => %Q["{{getv "/#{keys.join(".")}"}}"]})
            else
              output.merge!({k => %Q[{{getv "/#{keys.join(".")}"}}]})
            end
          end
        end
      end
      output
    end

    def recur_to_toml(input, current_keys, total_keys)
      if input.is_a?(Array)
        input.each_with_index do |v, k|
          if v.is_a?(Hash)  || v.is_a?(Array)
            recur_to_toml(v, current_keys + [k], total_keys)
          else
            keys = current_keys + [k]
            total_keys << "/#{keys.join(".")}" 
          end
        end
      else
        input.each do |k, v|
          if v.is_a?(Array) && terminal_arry?(v)
            keys = current_keys + [k]
            total_keys << "/#{keys.join(".")}" 
          elsif v.is_a?(Hash) || v.is_a?(Array)
            recur_to_toml(v, current_keys + [k], total_keys)
          else
            keys = current_keys + [k]
            total_keys << "/#{keys.join(".")}" 
          end
        end
      end
      total_keys
    end
  end
end
