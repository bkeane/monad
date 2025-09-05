<script setup>
import { CodeBlock } from 'vue-tufte';
</script>

<template>
    <section id="requirements">
        <h2>Requirements</h2>
        <p>Lambda functions require IAM roles and policies for AWS service access.</p>
    </section>
    
    <section id="role">
        <h2>Role</h2>
        <p>Custom IAM role trust policy:</p>
        <CodeBlock language="bash">
{{`monad deploy --role ./custom-role.json.tmpl`}}
        </CodeBlock>
    </section>
    
    <section id="policy">
        <h2>Policy</h2>
        <p>Custom IAM policy template:</p>
        <CodeBlock language="bash">
{{`monad deploy --policy ./custom-policy.json.tmpl`}}
        </CodeBlock>
    </section>
    
    <section id="boundary">
        <h2>Boundary</h2>
        <p>Optional permissions boundary for role:</p>
        <CodeBlock language="bash">
{{`monad deploy --boundary DeveloperBoundary`}}
        </CodeBlock>
    </section>
    
    <section id="convention">
        <h2>Convention</h2>
        
        <h3>Role</h3>
        <p>The default role allows Lambda service to assume it.</p>
        
        <h3>Policy</h3>
        <p>The default policy grants full CloudWatch Logs access for the function's log group.</p>
        
        <h3>Boundary</h3>
        <p>No permissions boundary is applied by default.</p>
    </section>
    
    <section id="configuration">
        <h2>Configuration</h2>
        
        <h3>Role</h3>
        <p>At init time you can generate the default role template for modification:</p>
        <CodeBlock language="bash">
{{`monad init go --role # output: role.json.tmpl
monad deploy --role ./role.json.tmpl`}}
        </CodeBlock>
        <p>...Or you can create a valid role json by hand.</p>
        
        <h3>Policy</h3>
        <p>At init time you can generate the default policy template for modification:</p>
        <CodeBlock language="bash">
{{`monad init python --policy # output: policy.json.tmpl
monad deploy --policy ./policy.json.tmpl`}}
        </CodeBlock>
        <p>...Or you can create a valid policy json by hand.</p>
        
        <h3>Boundary</h3>
        <p>Apply existing permissions boundary:</p>
        <CodeBlock language="bash">
{{`monad deploy --boundary DeveloperBoundary`}}
        </CodeBlock>
    </section>
</template>