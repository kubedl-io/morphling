export default [
  {
    path: '/',
    component: '../layouts/BlankLayout',
    routes: [
      {
        path: '/user',
        component: '../layouts/UserLayout',
        routes: [
          {
            name: 'login',
            path: '/user/login',
            component: './User/login',
          },
        ],
      },
      {
        path: '/',
        component: '../layouts/SecurityLayout',
        routes: [
          {
            path: '/',
            component: '../layouts/BasicLayout',
            authority: ['admin', 'user'],
            routes: [
              {
                path: '/',
                redirect: '/cluster',
              },
              // {
              //   path: '/welcome',
              //   name: 'welcome',
              //   icon: 'smile',
              //   component: './Welcome',
              // },
              {
                path: '/cluster',
                name: 'cluster',
                icon: 'home',
                component: './ClusterInfo',
              },
              {
                path: '/pe-monitor',
                name: 'pe-monitor',
                icon: 'unordered-list',
                component: './Experiments',
              },
              {
                path: '/pe-monitor/detail',
                component: './ExperimentDetail',
              },
              {
                path: '/pe-submit',
                name: 'pe-submit',
                icon: 'edit',
                // component: './JobSubmit',
                component: './ExperimentCreate',
              },
              {
                path: '/llm-service-version',
                name: 'llm-service-version',
                icon: 'dashboard',
                component: './LLMServiceVersion',
              },
              {
                path: '/admin',
                name: 'admin',
                icon: 'crown',
                component: './Admin',
                authority: ['admin'],
                routes: [
                  {
                    path: '/admin/sub-page',
                    name: 'sub-page',
                    icon: 'smile',
                    component: './ClusterInfo',
                    authority: ['admin'],
                  },
                ],
              },
              {
                component: './404',
              },
            ],
          },
          {
            component: './404',
          },
        ],
      },
    ],
  },
  {
    component: './404',
  },
];
