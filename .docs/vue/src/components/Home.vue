<template>
    <h3>Three Factor Deployments</h3>
    <p>
        Monad uses convention to reduce the number of factors necessary for deploying an application
        from <a href="https://12factor.net/">twelve</a> to three.
    </p>
    The remaining factors are:
    <ol>
        <li>
            <strong>Codebase</strong> - A git repository.
        </li>
        <li>
            <strong>Dependencies</strong> - ECR image/s.
        </li>
        <li>
            <strong>Configuration</strong> - Environment variables.
        </li>
    </ol>
</template>