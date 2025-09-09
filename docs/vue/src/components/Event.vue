<script setup>
import { CodeBlock, Figure, NewThought, Sidenote } from 'vue-tufte';
import worker from '../../assets/diagrams/deployment-event-1.png';
</script>

<template>
    <section id="about">
        <h2>About Eventing</h2>
        <p>
            <NewThought>EventBridge as a Pub/Sub service</NewThought> allows us to execute our code based on events
            beyond client http requests. Monad provides an easy and tractible way to wire up code to
            AWS EventBridge Enabling...
        </p>
        <ol>
            <li>Cron jobs<Sidenote><a href="https://docs.aws.amazon.com/sagemaker/latest/dg/model-monitor-schedule-expression.html">cron expression</a></Sidenote><Sidenote><a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/using-eventbridge-scheduler.html#using-eventbridge-scheduler-create">event scheduler</a></Sidenote>.</li>
            <li>Invoke based on AWS events<Sidenote><a href="https://docs.aws.amazon.com/eventbridge/latest/ref/events.html">aws event index</a></Sidenote>.</li>
            <li>Arbitrary pub/sub<Sidenote><a href="https://docs.aws.amazon.com/cli/latest/reference/events/put-events.html">put events</a></Sidenote>.</li>
            <li>Third part integrations<Sidenote><a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-saas.html">event partners</a></Sidenote>.</li>
            <li>Eventing to external APIs<Sidenote><a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-api-destinations.html">api destinations</a></Sidenote></li>
        </ol>
    </section>
    
    <section>
        <Figure :src="worker"></Figure>
    </section>

    <section id="bus">
        <h2>Bus</h2>
        <p>The bus flag wires the lambda up to a named bus<Sidenote><a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-bus.html">eventbridge bus</a></Sidenote></p>
        <CodeBlock language="bash">
{{`monad deploy --bus $bus`}}
        </CodeBlock>
    </section>
    
    <section id="rule">
        <h2>Rule</h2>
        <p>By default the lambda will be invoked by events matching a unicast pattern...</p>
        <CodeBlock language="json">
{{`{
  "source": [{
    "prefix": {
      "equals-ignore-case": "/\{\{.Git.Repo\}\}/\{\{.Git.Branch\}\}"
    }
  }],
  "detail": {
    "destination": [ { "equals-ignore-case": "/\{\{.Git.Repo\}\}/\{\{.Git.Branch\}\}/\{\{.Service.Name\}\}" } ]
  }
}`}}
        </CodeBlock>
                <p>... roughly translated:</p>
        <ul>
            <li>source: any service within this repository & branch</li>
            <li>destination: this service on this branch</li>
        </ul>
        <p>The rule flag can be used to customize event matching rules<Sidenote><a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-rules.html">eventbridge rule</a></Sidenote><Sidenote><a href="https://docs.aws.amazon.com/eventbridge/latest/userguide/eb-event-pattern-sandbox.html">eventBridge sandbox</a></Sidenote>:</p>
        <CodeBlock language="bash">
{{`monad deploy --bus $bus --rule ./rule.json.tmpl`}}
        </CodeBlock>
        <h3>Example</h3>
        <p>Let's say you wanted broadcast behavior instead of the afformentioned unicast behavior</p>
        <CodeBlock language="json">
{{`{
  "source": [{
    "anything-but": "/\{\{.Git.Repo\}\}/\{\{.Git.Branch\}\}/\{\{.Service.Name\}\}"
  }],
  "detail": {
    "destination": [{
      "wildcard": "/\{\{.Git.Repo\}\}/\{\{.Git.Branch\}\}/*"
    }]
  }
}
`}}
        </CodeBlock>
                <p>... roughly translated:</p>
        <ul>
            <li>source: any service except for itself</li>
            <li>destination: any service within this repository & branch</li>
        </ul>
    </section>
    
    <section id="configuration">
        <h2>Handling Events</h2>
        <p>
            So long as you have kept with the conventional artifact setup, specifically the usage of
            the AWS Lambda Web Adpater
            <Sidenote>
                <a href="https://github.com/awslabs/aws-lambda-web-adapter">
                    AWS Lambda Web Adapter
                </a>
            </Sidenote>, event types other than HTTP will be `POST`ed to `/events` in
            your webserver for handling. The body of said request will contain the full event.
        </p>
        <p>
            If you have strayed from convention, just implement your software as you would for any other
            EventBridge invoked Lambda via handler func.
        </p>
        <p>
            See the Lambda Web Adapter docs to customize the route to which events are forwarded.
        </p>
    </section>
</template>