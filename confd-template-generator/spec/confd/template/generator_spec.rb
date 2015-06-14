require 'spec_helper'

describe Confd::TemplateGenerator do
  it 'has a version number' do
    expect(Confd::TemplateGenerator::VERSION).not_to be nil
  end

  it 'initializes the generator with input' do
    hash = {"a" => "abc"}
    expect(Confd::TemplateGenerator.new(hash).input).to eq(hash)
  end
  
  it 'generates tmpl from the input hashmap' do
    hash = {"a" => {"b" => "abc"}}
    expect(Confd::TemplateGenerator.new(hash).to_tmpl).to eq({"a" => {"b" => "{{getv \"/a.b\"}}"}})
  end

  it 'generates tmpl from the complex input hashmap' do
    hash = {"a" => {"b" => "abc", "d" => "dc"}, "c" => 4}
    expect(Confd::TemplateGenerator.new(hash).to_tmpl).to eq({"a" => 
                                                                    {"b" => "{{getv \"/a.b\"}}",
                                                                    "d" => "{{getv \"/a.d\"}}"},
                                                                     "c" => "{{getv \"/c\"}}"})
  end

  it 'extract keys from the complex input hashmap' do
    hash = {"a" => {"b" => ["abc", "nyc"], "d" => "dc"}, "c" => 4}
    expect(Confd::TemplateGenerator.new(hash).to_toml).to eq(["/a.b" , "/a.d", "/c"])
  end

  it 'generates JSON from the complex input hashmap' do
    hash = {"a" => {"b" => "abc", "d" => "dc"}, "c" => 4}
    output = {"a.b" => "abc", "a.d" => "dc", "c" => 4}.to_json
    expect(Confd::TemplateGenerator.new(hash).to_json).to eq(output)
  end
end
