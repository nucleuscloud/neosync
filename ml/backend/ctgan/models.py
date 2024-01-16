import torch
from torch.nn import (BatchNorm1d, Dropout, LeakyReLU, Linear, Module, ReLU,
                      Sequential)


class Discriminator(Module):

    def calc_gradient_penalty(self, real_data, fake_data, device='cpu', pac=10, lambda_=10):

        alpha = torch.rand(real_data.size(0) // pac, 1, 1, device=device)
        alpha = alpha.repeat(1, pac, real_data.size(1))
        alpha = alpha.view(-1, real_data.size(1))

        interpolates = alpha * real_data + ((1 - alpha) * fake_data)

        disc_interpolates = self(interpolates)

        gradients = torch.autograd.grad(
            outputs=disc_interpolates, inputs=interpolates,
            grad_outputs=torch.ones(disc_interpolates.size(), device=device),
            create_graph=True, retain_graph=True, only_inputs=True
        )[0]

        gradient_penalty = ((
            gradients.view(-1, pac * real_data.size(1)).norm(2, dim=1) - 1
        ) ** 2).mean() * lambda_

        return gradient_penalty

    def __init__(self, input_dim, dis_dims, pack=10):
        super(Discriminator, self).__init__()
        dim = input_dim * pack
        self.pack = pack
        self.packdim = dim
        seq = []
        for item in list(dis_dims):
            seq += [Linear(dim, item), LeakyReLU(0.2), Dropout(0.5)]
            dim = item

        seq += [Linear(dim, 1)]
        self.seq = Sequential(*seq)

    def forward(self, input):
        assert input.size()[0] % self.pack == 0
        return self.seq(input.view(-1, self.packdim))


class Residual(Module):
    def __init__(self, i, o):
        super(Residual, self).__init__()
        self.fc = Linear(i, o)
        self.bn = BatchNorm1d(o)
        self.relu = ReLU()

    def forward(self, input):
        out = self.fc(input)
        out = self.bn(out)
        out = self.relu(out)
        return torch.cat([out, input], dim=1)


class Generator(Module):
    def __init__(self, embedding_dim, gen_dims, data_dim):
        super(Generator, self).__init__()
        dim = embedding_dim
        seq = []
        for item in list(gen_dims):
            seq += [Residual(dim, item)]
            dim += item
        seq.append(Linear(dim, data_dim))
        self.seq = Sequential(*seq)

    def forward(self, input):
        data = self.seq(input)
        return data
