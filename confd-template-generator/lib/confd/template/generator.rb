require "confd/template/generator/version"

module Confd
  class TemplateGenerator
    attr_reader :input
    def initialize(input)
      @input = input
    end

    def generate_templ
    end
  end
end
