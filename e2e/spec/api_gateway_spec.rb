require 'spec_helper'

describe "api gateway" do
    if Monad.substrate.apigateway_enabled?
        describe apigatewayv2(Monad.substrate.gateway_name) do
            it { should exist }
            it { should have_route_key(Monad.expected.route_key).with_target(Monad.expected.function_arn) }
        end

        if Monad.substrate.aws_auth_enabled? || Monad.substrate.no_auth_enabled?
            describe "Sigv4" do
                it "authenticated GET #{Monad.expected.url}: return 200", retry: 10 do
                    response = sigv4_get_request(Monad.expected.url)
                    expect(response.code).to eq("200")
                end

                it "unauthenticated GET #{Monad.expected.url}: return 403", retry: 10 do
                    response = Net::HTTP.get_response(URI(Monad.expected.url))
                    expect(response.code).to eq("403")
                end
            end
        end

        if Monad.substrate.jwt_auth_enabled?
            describe "Bearer" do
                it "authenticated: return 200", retry: 10 do
                    response = bearer_get_request(Monad.expected.url)
                    expect(response.code).to eq("200")
                end

                it "unauthenticated: return 401", retry: 10 do
                    response = Net::HTTP.get_response(URI(Monad.expected.url))
                    expect(response.code).to eq("401")
                end
            end
        end
    else
        describe apigatewayv2('kaixo') do
            it { should exist }
            it { should_not have_route_key(Monad.expected.route_key).with_target(Monad.expected.function_arn) }
        end
    end
end