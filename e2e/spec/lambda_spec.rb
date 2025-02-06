require 'spec_helper'

describe "lambda" do
  describe lambda(Monad.expected.function_name) do
    it { should exist }
    its(:function_name) { should eq Monad.expected.function_name }
    its(:role) { should eq Monad.expected.role_arn }
    include_examples 'common tag tests', Monad.expected
    include_examples 'common lambda env tests', Monad.expected

    if Monad.substrate.vpc_enabled?
      its(:vpc_config) do
        expect(subject.vpc_config.subnet_ids).to eq Monad.expected.vpc_subnet_ids
      end

      its(:vpc_config) do
        expect(subject.vpc_config.security_group_ids).to eq Monad.expected.vpc_security_group_ids
      end
    else
      its(:vpc_config) do
        expect(subject.vpc_config.subnet_ids).to be_empty
      end

      its(:vpc_config) do
        expect(subject.vpc_config.security_group_ids).to be_empty
      end
    end
  end

  describe iam_role(Monad.expected.role_name) do
      it { should exist }
      its(:assume_role_policy_document) { should include("AssumeRole") }
      its(:assume_role_policy_document) { should include("lambda.amazonaws.com") }
      include_examples 'common tag tests', Monad.expected
  end

  describe iam_policy(Monad.expected.policy_name) do
      it { should exist }
      it { should be_attached_to_role(Monad.expected.role_arn) }
      its(:attachment_count) { should eq 1 }
      include_examples 'common tag tests', Monad.expected
  end
end
