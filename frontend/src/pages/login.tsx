import { useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Key } from 'lucide-react'
import { loginSchema, type LoginFormData } from '@/schemas/auth.schema'
import { authApi } from '@/lib/api'
import { setFormErrors, getErrorMessage } from '@/lib/form-utils'
import {
  Form,
  FormField,
  FormItem,
  FormLabel,
  FormControl,
  FormMessage,
} from '@/components/ui/form'

export function LoginPage() {
  const navigate = useNavigate()
  const form = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      apiKey: '',
    },
  })

  const onSubmit = async (data: LoginFormData) => {
    try {
      const result = await authApi.validate(data.apiKey)
      if (!result.valid) {
        form.setError('root', { message: 'Invalid API Key' })
        return
      }
      // Store the API key as auth token
      localStorage.setItem('auth_token', data.apiKey)
      navigate('/admin/dashboard')
    } catch (err) {
      const errorMessage = getErrorMessage(err)
      setFormErrors(err, form.setError, errorMessage)
    }
  }

  return (
    <div className="bg-background flex min-h-screen items-center justify-center px-4">
      <div className="w-full max-w-md space-y-8">
        <div className="text-center">
          <h1 className="text-foreground text-4xl font-bold tracking-tight">Claude Proxy</h1>
          <p className="text-muted-foreground mt-2 text-sm">Admin Panel</p>
        </div>

        <div className="border-border bg-card rounded-lg border p-8 shadow-lg">
          <h2 className="text-card-foreground mb-6 text-2xl font-semibold">Admin Sign In</h2>

          {form.formState.errors.root && (
            <div className="bg-destructive/10 text-destructive mb-4 rounded-md p-3 text-sm">
              {form.formState.errors.root.message}
            </div>
          )}

          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="apiKey"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>API Key</FormLabel>
                    <FormControl>
                      <div className="relative">
                        <Key className="text-muted-foreground absolute top-1/2 left-3 h-5 w-5 -translate-y-1/2" />
                        <input
                          type="password"
                          placeholder="Enter your API key"
                          className="border-input bg-background text-foreground placeholder:text-muted-foreground focus:border-ring focus:ring-ring w-full rounded-md border py-2 pr-3 pl-10 focus:ring-2 focus:outline-none"
                          {...field}
                        />
                      </div>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />

              <button
                type="submit"
                disabled={form.formState.isSubmitting}
                className="bg-primary text-primary-foreground hover:bg-primary/90 w-full rounded-md px-4 py-2 font-medium transition-colors disabled:opacity-50"
              >
                {form.formState.isSubmitting ? 'Validating...' : 'Sign in'}
              </button>
            </form>
          </Form>

          <div className="text-muted-foreground mt-4 text-center text-xs">
            <p>Enter your API key to access the admin panel</p>
          </div>
        </div>
      </div>
    </div>
  )
}
