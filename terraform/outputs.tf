locals {
  common = {
    runs-on = "ubuntu-latest"
    permissions = {
      id-token = "write"
      contents = "read"
    }
  }

  publish = [
    for path, service in var.services : {
      name = "Publish ${basename(path)}"
      run  = "${trimspace("monad --chdir ${path} encode ${service.encode_args}")} | docker compose -f - build --push"
    }
  ]

  deploy = [
    for path, service in var.services : {
      name = "Deploy ${basename(path)}"
      run  = trimspace("monad --chdir ${path} deploy ${service.deploy_args}")
    }
  ]
}

output "deploy" {
  value = yamlencode({
    env = {
      MONAD_REGISTRY_ID     = local.ecr_hub_account_id
      MONAD_REGISTRY_REGION = local.ecr_hub_account_region
      MONAD_BRANCH          = "$${{ github.head_ref || github.ref_name }}"
      MONAD_SHA             = "$${{ github.event_name == 'pull_request' && github.event.pull_request.head.sha || github.sha }}"
    }

    jobs = {
      publish = merge(local.common, {
        steps = concat([
          {
            name = "Setup Monad"
            id   = "setup-monad"
            uses = "bkeane/monad-action@main"
            with = {
              version         = "latest"
              role_arn        = local.ecr_hub_account_role_arn
              registry_id     = "$${{ env.MONAD_REGISTRY_ID }}"
              registry_region = "$${{ env.MONAD_REGISTRY_REGION }}"
            }
          }
        ], local.publish)
      }),

      deploy = merge(local.common, {
        needs = "publish"
        strategy = {
          matrix = {
            role_arn = compact(flatten([
              local.ecr_hub_account_role_arn,
              local.ecr_spoke_account_role_arns
            ]))
          }
        }
        steps = concat([
          {
            name = "Setup Monad"
            id   = "setup-monad"
            uses = "bkeane/monad-action@main"
            with = {
              version         = "latest"
              role_arn        = "$${{ matrix.role_arn }}"
              registry_id     = "$${{ env.MONAD_REGISTRY_ID }}"
              registry_region = "$${{ env.MONAD_REGISTRY_REGION }}"
            }
          }
        ], local.deploy)
      })
    }
  })
}
