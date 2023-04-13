/*
 * @Author: Bin
 * @Date: 2023-03-18
 * @FilePath: /gpt-zmide-server/src/pages/admin/screens/application/index.tsx
 */
import React from 'react'
import { Button, Input, Message, Modal, Result, Table, TableColumnProps, Tag, Form, Tooltip, Dropdown, Menu } from '@arco-design/web-react'
import { IconQuestionCircle } from '@arco-design/web-react/icon'
import useAxios from 'axios-hooks';
import { axios } from '@/apis';

type createAppConfigType = {
    visible: boolean,
    id?: number,
    name?: string
}

export default function index() {

    const [{ data, error, loading }, refresh] = useAxios({
        url: "/api/admin/application/"
    })

    const columns: TableColumnProps[] = [
        {
            title: '应用名称',
            dataIndex: 'name',
        },
        {
            title: '密钥',
            dataIndex: 'app_key',
        },
        {
            title: 'API_KEY',
            dataIndex: 'api_key',
        },
        {
            title: <Tooltip content='OpenAI 接口对 gpt-3.5 消息上下文有 4600 字数限制，启用修复长消息的话，当会话消息字数超过限制会自动忽略旧消息，只发送最新消息内容'>
                长消息<IconQuestionCircle />
            </Tooltip>,
            dataIndex: 'enable_fix_long_msg',
            render: (enable_fix_long_msg) => {
                return enable_fix_long_msg === 1 ? <Tag color='green'>启用</Tag> : <Tag color='red'>禁用</Tag>
            }
        },
        {
            title: '状态',
            dataIndex: 'status',
            render: (status) => {
                return status === 1 ? <Tag color='green'>已启用</Tag> : <Tag color='red'>已禁用</Tag>
            }
        },
        {
            title: '创建时间',
            dataIndex: 'created_at',
        },
        {
            title: '操作',
            dataIndex: 'id',
            align: 'center',
            render: (id, item) => {
                return item ? <>
                    <Button
                        type='text'
                        disabled={!item?.id}
                        onClick={() => {
                            setCreateAppConfig({
                                ...createAppConfig,
                                visible: true,
                                id: item?.id,
                                name: item?.name,
                            })
                        }}
                    >
                        修改
                    </Button>
                    <Dropdown
                        droplist={
                            <Menu onClickMenuItem={(key) => {
                                if (key == 'status') {
                                    return updateAppStatus(item.id, item.status)
                                }

                                if (key == 'long_message') {
                                    return updateAppStatus(item.id, undefined, item.enable_fix_long_msg)
                                }

                                if (key == 'reset_apikey') {
                                    return resetAppApiKey(item.id)
                                }
                            }}>
                                <Menu.Item key='status'>{item?.status === 1 ? '禁用' : '启用'}</Menu.Item>
                                <Menu.Item key='long_message'>{item?.enable_fix_long_msg === 1 ? '禁用/长消息' : '启用/长消息'}</Menu.Item>
                                <Menu.Item key='reset_apikey'>重置API_KEY</Menu.Item>
                            </Menu>
                        }
                    >
                        <Button type='text'>更多</Button>
                    </Dropdown>
                </> : undefined
            }
        },
    ];

    const [createAppConfig, setCreateAppConfig] = React.useState<createAppConfigType>({
        visible: false,
        id: undefined,
        name: undefined,
    })

    // 创建应用
    const createApp = (name: string) => {
        if (!name || name == "") {
            Message.warning('应用名不得为空。')
            return
        }

        const formData = new FormData();
        formData.append("name", name)

        axios.post("/api/admin/application/create", formData).then((response) => {
            const { code, msg, data } = response.data
            if (code !== 200) {
                Message.info(`请求失败，${msg || code}`)
                return
            }
            // 成功
            refresh() // 刷新数据 
            setCreateAppConfig({
                ...createAppConfig,
                visible: false,
                id: undefined,
                name: undefined,
            }) // 关闭弹窗
        }).catch(err => {
            Message.info(`请求失败，${err.message || '请稍后重试'}`)
        })
    }

    // 更新应用
    const updateApp = (id: number, name: string) => {
        if (!name || name == "") {
            Message.warning('应用名不得为空。')
            return
        }

        const formData = new FormData();
        formData.append("name", name)

        axios.post(`/api/admin/application/${id}/update`, formData).then((response) => {
            const { code, msg, data } = response.data
            if (code !== 200) {
                Message.info(`请求失败，${msg || code}`)
                return
            }
            // 成功
            refresh() // 刷新数据 
            setCreateAppConfig({
                ...createAppConfig,
                visible: false,
                id: undefined,
                name: undefined,
            }) // 关闭弹窗
        }).catch(err => {
            Message.info(`请求失败，${err.message || '请稍后重试'}`)
        })
    }

    // 更新应用状态
    const updateAppStatus = (id: number, status?: number, fix_long_msg?: number) => {
        if (!id || id < 1) {
            Message.warning('应用异常。')
            return
        }

        const formData = new FormData();
        if (status !== undefined) {
            formData.append("status", status === 1 ? '2' : '1')
        }
        if (fix_long_msg != undefined) {
            formData.append("fix_long_msg", fix_long_msg === 1 ? '2' : '1')
        }

        axios.post(`/api/admin/application/${id}/update`, formData).then((response) => {
            const { code, msg, data } = response.data
            if (code !== 200) {
                Message.info(`请求失败，${msg || code}`)
                return
            }
            // 成功
            refresh() // 刷新数据 
            Message.success(`配置成功`)
        }).catch(err => {
            Message.info(`请求失败，${err.message || '请稍后重试'}`)
        })
    }

    // 重置 api_key
    const resetAppApiKey = (id: number) => {
        if (!id || id < 1) {
            Message.warning('应用异常。')
            return
        }

        const formData = new FormData();

        axios.post(`/api/admin/application/${id}/apikey/reset`, formData).then((response) => {
            const { code, msg, data } = response.data
            if (code !== 200) {
                Message.info(`请求失败，${msg || code}`)
                return
            }
            // 成功
            refresh() // 刷新数据 
            Message.success(`操作成功`)
        }).catch(err => {
            Message.info(`请求失败，${err.message || '请稍后重试'}`)
        })
    }

    return (
        <div style={{ marginTop: 20 }}>
            <div style={{ display: 'flex', flexDirection: 'row' }}>
                <div style={{ flex: 1 }}></div>
                <Button
                    style={{ marginBottom: 10, }}
                    type='primary'
                    onClick={() => setCreateAppConfig({
                        ...createAppConfig,
                        visible: true,
                        id: undefined,
                        name: undefined,
                    })}
                >
                    创建应用
                </Button>
            </div>
            {loading || (!error && data?.data) ? (
                <Table rowKey={(item) => 'app_item_' + item.id} loading={loading} columns={columns} data={data?.data} />
            ) : (
                <Result
                    status='warning'
                    title={error ? `出错啦，${error.message}` : '应用列表为空，请先创建一个应用。'}
                    extra={<Button type='primary' onClick={refresh}>刷新</Button>}
                />
            )}

            <Modal
                title={createAppConfig.id ? '修改应用' : '创建新应用'}
                visible={createAppConfig.visible}
                onOk={() => {
                    // console.log("创建应用", createAppConfig?.name);
                    if (createAppConfig?.id) {
                        // 修改应用
                        updateApp(createAppConfig.id, createAppConfig.name || '')
                    } else {
                        // 创建应用
                        createApp(createAppConfig?.name || '')
                    }
                }}
                onCancel={() => setCreateAppConfig({
                    ...createAppConfig,
                    visible: false,
                })}
                autoFocus={false}
                focusLock={true}
                okText={createAppConfig.id ? '修改' : '创建'}
            >
                {createAppConfig.visible && (
                    <Form autoComplete='off' layout="vertical" >
                        <Form.Item label='应用名称'>
                            <Input defaultValue={createAppConfig.name} onChange={(value) => {
                                setCreateAppConfig({
                                    ...createAppConfig,
                                    name: value,
                                })
                            }} placeholder='请输入应用名称' />
                        </Form.Item>
                    </Form>
                )}
            </Modal>
        </div>
    )
}
