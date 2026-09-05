import van from 'vanjs-core'
import { login } from '../api/admin'
import { ApiError } from '../api/http'
import { setToken } from '../store/auth'
import { fetchMenus } from '../store/menu'
import { navigate, HOME_PATH } from '../router/router'
import { toast } from '../components/Toast'
import { C, fieldInput, fieldLabel } from '../ui/kit'

const { div, form, h2, p, button, img } = van.tags

/** Login page: username + password, JWT into localStorage. */
export function LoginView(): HTMLElement {
  const submitting = van.state(false)

  const onSubmit = async (e: Event): Promise<void> => {
    e.preventDefault()
    if (submitting.val) return
    const formEl = e.currentTarget as HTMLFormElement
    const data = new FormData(formEl)
    const username = String(data.get('username') ?? '').trim()
    const password = String(data.get('password') ?? '')
    if (!username || !password) {
      toast.error('请输入用户名和密码')
      return
    }
    submitting.val = true
    try {
      const res = await login({ username, password })
      setToken(res.accessToken)
      toast.success('登录成功')
      void fetchMenus().catch(() => undefined)
      navigate(HOME_PATH)
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : '登录失败，请稍后重试'
      toast.error(msg)
    } finally {
      submitting.val = false
    }
  }

  return div(
    {
      style:
        'min-height: 100vh; display: flex; align-items: center; justify-content: center; background: linear-gradient(135deg, #eef4f8 0%, #e3edf5 100%); padding: 16px;'
    },
    div(
      {
        style: `width: min(380px, 100%); background: #fff; border: 1px solid ${C.border}; border-radius: 14px; box-shadow: 0 10px 40px rgba(16, 149, 193, 0.12); padding: 32px 35px 35px; text-align: center;`
      },
      img({ src: '/favicon.svg', alt: 'Nunu Admin', style: 'width: 46px; height: 46px;' }),
      h2({ style: 'margin: 10px 0 3px;' }, 'Nunu Admin'),
      p({ style: `color: #7c8ca0; margin: 0 0 22px; font-size: 0.88rem;` }, 'VanJS + VanUI 管理后台'),
      form(
        { onsubmit: onSubmit, style: 'text-align: left;' },
        fieldLabel(
          '用户名',
          fieldInput({ type: 'text', name: 'username', placeholder: 'admin', autocomplete: 'username', required: true })
        ),
        fieldLabel(
          '密码',
          fieldInput({
            type: 'password',
            name: 'password',
            placeholder: '••••••',
            autocomplete: 'current-password',
            required: true
          })
        ),
        button(
          {
            type: 'submit',
            disabled: (): boolean => submitting.val,
            'aria-busy': (): boolean => submitting.val,
            style: `width: 100%; margin-top: 6px; padding: 10px 16px; font-size: 0.98rem; font-family: inherit; border-radius: 8px; border: 1px solid ${C.brand}; background: ${C.brand}; color: #fff; cursor: pointer;`
          },
          submitting.val ? '登录中…' : '登 录'
        )
      )
    )
  )
}
