import { defineConfig } from 'vitepress';

export default defineConfig({
  title: "Documentation",
  lang: 'en-US',
  base: '/',
  lastUpdated: true,
  cleanUrls: true,

  head: [
    ['link', { rel: 'icon', href: '/favicon.ico' }]
  ],

  themeConfig: {
    nav: [
    {
        "text": "Getting Started",
        "link": "/getting-started/index"
    },
    {
        "text": "Core Operations",
        "link": "/core-operations/index"
    },
    {
        "text": "Task-Based Guide",
        "link": "/task-based-guide/index"
    },
    {
        "text": "Advanced Usage",
        "link": "/advanced-usage/index"
    },
    {
        "text": "Problem Solving",
        "link": "/problem-solving/index"
    },
    {
        "text": "Reference",
        "link": "/reference/system"
    }
],
    sidebar: [
    {
        "text": "Getting Started",
        "items": [
            {
                "text": "Overview",
                "link": "/getting-started/index#overview"
            },
            {
                "text": "Core Concepts",
                "link": "/getting-started/index#core-concepts"
            },
            {
                "text": "Quick Setup Guide",
                "link": "/getting-started/index#quick-setup-guide"
            },
            {
                "text": "Prerequisites",
                "link": "/getting-started/index#prerequisites"
            },
            {
                "text": "Installation Steps",
                "link": "/getting-started/index#installation-steps"
            },
            {
                "text": "Verification",
                "link": "/getting-started/index#verification"
            }
        ]
    },
    {
        "text": "Core Operations",
        "items": [
            {
                "text": "Queueing Tasks",
                "link": "/core-operations/index#queueing-tasks"
            },
            {
                "text": "Immediate Task Execution (`RunTask`)",
                "link": "/core-operations/index#immediate-task-execution-runtask"
            }
        ]
    },
    {
        "text": "Task-Based Guide",
        "items": [
            {
                "text": "Resource Lifecycle Management",
                "link": "/task-based-guide/index#resource-lifecycle-management"
            },
            {
                "text": "Health Checks & Retries",
                "link": "/task-based-guide/index#health-checks-retries"
            },
            {
                "text": "Asynchronous Submission Patterns",
                "link": "/task-based-guide/index#asynchronous-submission-patterns"
            },
            {
                "text": "\"At-Most-Once\" Execution",
                "link": "/task-based-guide/index#at-most-once-execution"
            },
            {
                "text": "Dynamic Scaling (Bursting)",
                "link": "/task-based-guide/index#dynamic-scaling-bursting"
            },
            {
                "text": "Monitoring & Observability",
                "link": "/task-based-guide/index#monitoring-observability"
            }
        ]
    },
    {
        "text": "Advanced Usage",
        "items": [
            {
                "text": "Graceful Shutdown (`Stop`)",
                "link": "/advanced-usage/index#graceful-shutdown-stop"
            },
            {
                "text": "Immediate Shutdown (`Kill`)",
                "link": "/advanced-usage/index#immediate-shutdown-kill"
            }
        ]
    },
    {
        "text": "Problem Solving",
        "items": [
            {
                "text": "Troubleshooting Common Issues",
                "link": "/problem-solving/index#troubleshooting-common-issues"
            }
        ]
    },
    {
        "text": "Types",
        "items": [
            {
                "text": "Config",
                "link": "/reference/types#config"
            },
            {
                "text": "TaskStats",
                "link": "/reference/types#taskstats"
            },
            {
                "text": "TaskMetrics",
                "link": "/reference/types#taskmetrics"
            },
            {
                "text": "TaskLifecycleTimestamps",
                "link": "/reference/types#tasklifecycletimestamps"
            }
        ]
    },
    {
        "text": "Interfaces",
        "items": [
            {
                "text": "Logger",
                "link": "/reference/interfaces#logger"
            },
            {
                "text": "MetricsCollector",
                "link": "/reference/interfaces#metricscollector"
            },
            {
                "text": "TaskManager",
                "link": "/reference/interfaces#taskmanager"
            }
        ]
    },
    {
        "text": "Examples",
        "items": [
            {
                "text": "Basic tasker setup",
                "link": "/reference/patterns#basic-tasker-setup"
            },
            {
                "text": "Custom health check",
                "link": "/reference/patterns#custom-health-check"
            },
            {
                "text": "Graceful shutdown",
                "link": "/reference/patterns#graceful-shutdown"
            },
            {
                "text": "Immediate shutdown",
                "link": "/reference/patterns#immediate-shutdown"
            },
            {
                "text": "Get live stats",
                "link": "/reference/patterns#get-live-stats"
            },
            {
                "text": "Get performance metrics",
                "link": "/reference/patterns#get-performance-metrics"
            },
            {
                "text": "Queue and wait",
                "link": "/reference/patterns#queue-and-wait"
            },
            {
                "text": "Queue with callback",
                "link": "/reference/patterns#queue-with-callback"
            },
            {
                "text": "Queue asynchronously with channels",
                "link": "/reference/patterns#queue-asynchronously-with-channels"
            }
        ]
    },
    {
        "text": "Reference",
        "items": [
            {
                "text": "System Overview",
                "link": "/reference/system"
            },
            {
                "text": "Dependencies",
                "link": "/reference/dependencies"
            },
            {
                "text": "Integration",
                "link": "/reference/integration"
            },
            {
                "text": "Methods",
                "link": "/reference/methods"
            },
            {
                "text": "Error Reference",
                "link": "/reference/errors"
            }
        ]
    }
],

    socialLinks: [
      // { icon: 'github', link: '[https://github.com/asaidimu/$](https://github.com/asaidimu/$){data.reference.system.name.toLowerCase()}' }
    ],

    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2023-present tasker'
    }
  }
});