require 'spec_helper'

describe "eventbridge" do
    if Monad.substrate.eventbridge_enabled?
        describe eventbridge(Monad.expected.bus_name) do
            buses(Monad.expected.path) do |rule|
                it { should exist }
                it { should have_rule("#{Monad.substrate.resource_name}-#{rule}").with_target(Monad.expected.function_arn) }
            end
        end
    else
        describe eventbridge(Monad.expected.bus_name) do
            buses(Monad.expected.path) do |rule|
                it { should exist }
                it { should_not have_rule("#{Monad.substrate.resource_name}-#{rule}") }
            end
        end
    end
end