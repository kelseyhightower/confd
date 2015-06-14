require 'yaml'
require 'confd/template/generator'

module Confd
  class YamlGenerator
    attr_accessor :name, :confd_generator, :bucket, :dest
    def initialize(input, name, bucket, dest)
      @name = name
      @confd_generator = TemplateGenerator.new(YAML.load_file(input))
      @bucket = bucket
      @dest = dest
    end

    def generate
      generate_tmpl
      generate_toml
      generate_json
    end

    private

    def generate_tmpl
      File.open("#{Dir.pwd}/#{name}.tmpl", "w") { |f| f.write(sanatized_tmpl) }
    end

    def generate_toml
      toml = %{[template]\nprefix="/#{bucket}"\nsrc = "#{name}.tmpl"\ndest = "#{dest}"\nkeys = #{confd_generator.to_toml}}
      File.open("#{Dir.pwd}/#{name}.toml", "w") { |f| f.write(toml) }
    end

    def generate_json
      File.open("#{Dir.pwd}/#{name}.json", "w") { |f| f.write(confd_generator.to_json) }
    end

    def sanatized_tmpl
      confd_generator.to_tmpl.to_yaml.gsub(": |", ":").gsub("'{{", "{{").gsub("}}'", "}}")
    end
  end
end
