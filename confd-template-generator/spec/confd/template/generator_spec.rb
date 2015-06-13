require 'spec_helper'

describe Confd::TemplateGenerator do
  it 'has a version number' do
    expect(Confd::TemplateGenerator::VERSION).not_to be nil
  end

  it 'initializes the generator with input' do
    hash = {"a" => "abc"}
    expect(Confd::TemplateGenerator.new(hash).input).to eq(hash)
  end
  
  it 'generates templ from the input hashmap' do
    hash = {"a" => "abc"}
    expect(Confd::TemplateGenerator.new(hash).input).to eq(hash)
  end
end
